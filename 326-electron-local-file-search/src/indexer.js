const fs = require('fs');
const path = require('path');
const crypto = require('crypto');
const EventEmitter = require('events');

const MAX_FILE_SIZE = 5 * 1024 * 1024;
const MAX_LINES_PER_LARGE_FILE = 1000;
const CACHE_DIR = path.join(require('os').homedir(), '.local-file-search', 'cache');

const DEFAULT_EXTENSIONS = [
  '.txt', '.md', '.js', '.py', '.json', '.html', '.css', '.xml', '.log', '.csv'
];

class FileIndexer extends EventEmitter {
  constructor() {
    super();
    this.index = null;
    this.currentDirectory = null;
    this.extensions = [...DEFAULT_EXTENSIONS];
    this._ensureCacheDir();
  }

  _ensureCacheDir() {
    if (!fs.existsSync(CACHE_DIR)) {
      fs.mkdirSync(CACHE_DIR, { recursive: true });
    }
  }

  _getCacheKey(directory) {
    return crypto.createHash('sha256').update(directory).digest('hex');
  }

  _getCachePath(directory) {
    const cacheKey = this._getCacheKey(directory);
    return path.join(CACHE_DIR, `${cacheKey}.json`);
  }

  hasCache(directory) {
    const cachePath = this._getCachePath(directory);
    return fs.existsSync(cachePath);
  }

  loadCache(directory) {
    const cachePath = this._getCachePath(directory);
    if (!fs.existsSync(cachePath)) {
      return null;
    }
    
    const cacheData = JSON.parse(fs.readFileSync(cachePath, 'utf-8'));
    this.index = cacheData.index;
    this.currentDirectory = directory;
    this.extensions = cacheData.extensions || [...DEFAULT_EXTENSIONS];
    return cacheData;
  }

  _saveCache(directory, indexData) {
    const cachePath = this._getCachePath(directory);
    const cacheData = {
      directory,
      extensions: this.extensions,
      timestamp: Date.now(),
      index: indexData
    };
    fs.writeFileSync(cachePath, JSON.stringify(cacheData));
  }

  clearCache() {
    if (this.currentDirectory) {
      const cachePath = this._getCachePath(this.currentDirectory);
      if (fs.existsSync(cachePath)) {
        fs.unlinkSync(cachePath);
      }
    }
    this.index = null;
  }

  _shouldIndexFile(filePath) {
    const ext = path.extname(filePath).toLowerCase();
    return this.extensions.includes(ext);
  }

  _listFiles(directory) {
    const files = [];
    
    const walkDir = (dir) => {
      const entries = fs.readdirSync(dir, { withFileTypes: true });
      
      for (const entry of entries) {
        const fullPath = path.join(dir, entry.name);
        
        if (entry.isDirectory()) {
          if (entry.name !== 'node_modules' && 
              entry.name !== '.git' && 
              entry.name !== 'dist' &&
              entry.name !== 'build') {
            walkDir(fullPath);
          }
        } else if (entry.isFile()) {
          if (this._shouldIndexFile(fullPath)) {
            files.push(fullPath);
          }
        }
      }
    };
    
    walkDir(directory);
    return files;
  }

  _readFileLines(filePath) {
    const stats = fs.statSync(filePath);
    const isLargeFile = stats.size > MAX_FILE_SIZE;
    
    try {
      const content = fs.readFileSync(filePath, 'utf-8');
      let lines = content.split(/\r?\n/);
      
      if (isLargeFile && lines.length > MAX_LINES_PER_LARGE_FILE) {
        lines = lines.slice(0, MAX_LINES_PER_LARGE_FILE);
      }
      
      return lines;
    } catch (error) {
      return [];
    }
  }

  _tokenize(text) {
    return text.toLowerCase().split(/[\s,.\-_!?;:()\[\]{}"'\/\\]+/).filter(t => t.length > 0);
  }

  async indexDirectory(directory, extensions = null) {
    if (extensions) {
      this.extensions = extensions.map(ext => 
        ext.startsWith('.') ? ext.toLowerCase() : `.${ext.toLowerCase()}`
      );
    }
    
    this.currentDirectory = directory;
    
    const files = this._listFiles(directory);
    const totalFiles = files.length;
    let processedFiles = 0;
    
    const invertedIndex = {};
    const fileDocs = {};
    
    for (const filePath of files) {
      this.emit('file-scanning', filePath);
      
      const lines = this._readFileLines(filePath);
      const docId = filePath;
      
      const terms = {};
      
      lines.forEach((line, lineNum) => {
        const tokens = this._tokenize(line);
        tokens.forEach(token => {
          if (!terms[token]) {
            terms[token] = [];
          }
          if (!terms[token].includes(lineNum)) {
            terms[token].push(lineNum);
          }
          
          if (!invertedIndex[token]) {
            invertedIndex[token] = {};
          }
          if (!invertedIndex[token][docId]) {
            invertedIndex[token][docId] = [];
          }
          if (!invertedIndex[token][docId].includes(lineNum)) {
            invertedIndex[token][docId].push(lineNum);
          }
        });
      });
      
      fileDocs[docId] = {
        path: filePath,
        lines: lines,
        terms: Object.keys(terms)
      };
      
      processedFiles++;
      const progress = {
        current: processedFiles,
        total: totalFiles,
        percent: Math.round((processedFiles / totalFiles) * 100)
      };
      this.emit('progress', progress);
    }
    
    this.index = {
      invertedIndex,
      fileDocs
    };
    
    this._saveCache(directory, this.index);
    
    return {
      totalFiles,
      directory,
      extensions: this.extensions,
      timestamp: Date.now()
    };
  }

  _parseQuery(query) {
    const tokens = query.trim().split(/\s+/);
    const must = [];
    const should = [];
    const mustNot = [];
    
    let currentGroup = must;
    
    for (const token of tokens) {
      if (token === '|' || token === 'OR') {
        currentGroup = should;
        continue;
      }
      
      if (token.startsWith('-')) {
        const term = token.slice(1).toLowerCase();
        if (term) mustNot.push(term);
      } else {
        const term = token.toLowerCase();
        if (term) currentGroup.push(term);
      }
    }
    
    return { must, should, mustNot };
  }

  _getDocsForTerm(term) {
    if (!this.index || !this.index.invertedIndex) return {};
    return this.index.invertedIndex[term] || {};
  }

  search(query, options = {}) {
    if (!this.index || !this.index.invertedIndex) {
      return [];
    }
    
    const { must, should, mustNot } = this._parseQuery(query);
    
    if (must.length === 0 && should.length === 0) {
      return [];
    }
    
    let resultDocs = null;
    
    if (must.length > 0) {
      resultDocs = this._getDocsForTerm(must[0]);
      
      for (let i = 1; i < must.length; i++) {
        const termDocs = this._getDocsForTerm(must[i]);
        const newDocs = {};
        
        for (const docId in resultDocs) {
          if (termDocs[docId]) {
            newDocs[docId] = [...resultDocs[docId]];
            for (const line of termDocs[docId]) {
              if (!newDocs[docId].includes(line)) {
                newDocs[docId].push(line);
              }
            }
            newDocs[docId].sort((a, b) => a - b);
          }
        }
        resultDocs = newDocs;
        
        if (Object.keys(resultDocs).length === 0) break;
      }
    }
    
    if (should.length > 0) {
      const shouldDocs = {};
      
      for (const term of should) {
        const termDocs = this._getDocsForTerm(term);
        for (const docId in termDocs) {
          if (!shouldDocs[docId]) {
            shouldDocs[docId] = [...termDocs[docId]];
          } else {
            for (const line of termDocs[docId]) {
              if (!shouldDocs[docId].includes(line)) {
                shouldDocs[docId].push(line);
              }
            }
            shouldDocs[docId].sort((a, b) => a - b);
          }
        }
      }
      
      if (resultDocs) {
        const mergedDocs = {};
        for (const docId in resultDocs) {
          mergedDocs[docId] = resultDocs[docId];
        }
        for (const docId in shouldDocs) {
          if (mergedDocs[docId]) {
            for (const line of shouldDocs[docId]) {
              if (!mergedDocs[docId].includes(line)) {
                mergedDocs[docId].push(line);
              }
            }
            mergedDocs[docId].sort((a, b) => a - b);
          } else {
            mergedDocs[docId] = shouldDocs[docId];
          }
        }
        resultDocs = mergedDocs;
      } else {
        resultDocs = shouldDocs;
      }
    }
    
    if (mustNot.length > 0 && resultDocs) {
      for (const term of mustNot) {
        const termDocs = this._getDocsForTerm(term);
        for (const docId in termDocs) {
          delete resultDocs[docId];
        }
      }
    }
    
    const allSearchTerms = [...must, ...should];
    const results = [];
    
    for (const docId in resultDocs) {
      const fileDoc = this.index.fileDocs[docId];
      if (!fileDoc) continue;
      
      const matchingLines = resultDocs[docId];
      const matches = matchingLines.map(lineNum => ({
        line: lineNum + 1,
        content: fileDoc.lines[lineNum] || ''
      }));
      
      results.push({
        filePath: fileDoc.path,
        matches,
        matchCount: matches.length,
        termMatches: allSearchTerms.length
      });
    }
    
    results.sort((a, b) => {
      if (b.termMatches !== a.termMatches) {
        return b.termMatches - a.termMatches;
      }
      return b.matchCount - a.matchCount;
    });
    
    return results;
  }

  getStats() {
    if (!this.index) {
      return { initialized: false };
    }
    
    return {
      initialized: true,
      directory: this.currentDirectory,
      totalFiles: Object.keys(this.index.fileDocs).length,
      totalTerms: Object.keys(this.index.invertedIndex).length,
      extensions: this.extensions
    };
  }
}

module.exports = FileIndexer;
