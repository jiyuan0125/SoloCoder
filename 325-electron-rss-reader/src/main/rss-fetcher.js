const Parser = require('rss-parser');
const https = require('https');
const http = require('http');
const { URL } = require('url');

class RssFetcher {
  constructor(dataStore) {
    this.dataStore = dataStore;
    this.parser = new Parser({
      timeout: 30000,
      maxRedirects: 5
    });
  }

  validateUrl(url) {
    try {
      new URL(url);
      return true;
    } catch {
      return false;
    }
  }

  async validateFeedUrl(url) {
    if (!this.validateUrl(url)) {
      return { valid: false, error: '无效的 URL 格式' };
    }

    try {
      const feed = await this.parser.parseURL(url);
      const isRssOrAtom = this.isRssOrAtom(feed);
      
      if (!isRssOrAtom) {
        return { valid: false, error: 'URL 不是有效的 RSS 或 Atom 格式' };
      }

      return { 
        valid: true, 
        title: feed.title || url,
        feed: feed 
      };
    } catch (error) {
      return { valid: false, error: this.getErrorMessage(error) };
    }
  }

  isRssOrAtom(feed) {
    return !!(feed.title || feed.items || feed.feedUrl);
  }

  getErrorMessage(error) {
    if (error.code === 'ENOTFOUND') {
      return '无法连接到服务器，请检查网络连接';
    }
    if (error.code === 'ECONNREFUSED') {
      return '连接被拒绝';
    }
    if (error.code === 'ETIMEDOUT') {
      return '连接超时';
    }
    if (error.message?.includes('status code')) {
      return '服务器返回错误响应';
    }
    return '无法解析 RSS/Atom 格式: ' + (error.message || '未知错误');
  }

  async fetchAndUpdateFeed(url) {
    const feeds = this.dataStore.getAllFeeds();
    const feed = feeds.find(f => f.url.toLowerCase() === url.toLowerCase());
    
    if (!feed) {
      throw new Error('订阅源不存在');
    }

    try {
      const parsedFeed = await this.parser.parseURL(url);
      
      if (parsedFeed.title && parsedFeed.title !== feed.title) {
        this.dataStore.updateFeed({ ...feed, title: parsedFeed.title });
      }

      const existingArticles = this.dataStore.getArticlesByFeed(feed.id);
      const existingGuids = new Set(existingArticles.map(a => a.guid));

      for (const item of parsedFeed.items) {
        const guid = item.guid || item.id || item.link;
        
        if (existingGuids.has(guid)) {
          continue;
        }

        const article = this.normalizeArticle(feed.id, item);
        this.dataStore.addArticle(article);
      }

      return { success: true };
    } catch (error) {
      console.error(`Failed to fetch feed ${url}:`, error.message);
      throw error;
    }
  }

  normalizeArticle(feedId, item) {
    return {
      feedId,
      title: item.title || '无标题',
      link: item.link || '',
      guid: item.guid || item.id || item.link,
      pubDate: item.pubDate || item.isoDate || new Date().toISOString(),
      author: item.creator || item.author || '',
      summary: this.stripHtml(item.contentSnippet || item.summary || ''),
      content: item.content || item['content:encoded'] || '',
      categories: item.categories || []
    };
  }

  stripHtml(html) {
    if (!html) return '';
    return html
      .replace(/<[^>]*>/g, '')
      .replace(/&nbsp;/g, ' ')
      .replace(/\s+/g, ' ')
      .trim();
  }
}

module.exports = RssFetcher;
