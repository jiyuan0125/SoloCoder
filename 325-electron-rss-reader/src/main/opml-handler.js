const { XMLParser, XMLBuilder } = require('fast-xml-parser');

class OpmlHandler {
  constructor() {
    this.parser = new XMLParser({
      ignoreAttributes: false,
      attributeNamePrefix: '@_',
      isArray: (name, jpath, isLeafNode, isAttribute) => {
        if (['outline', 'category'].includes(name)) return true;
        return false;
      }
    });
    this.builder = new XMLBuilder({
      ignoreAttributes: false,
      attributeNamePrefix: '@_',
      format: true
    });
  }

  exportOpml(feeds, categories) {
    const categoryMap = new Map();
    categories.forEach(cat => {
      categoryMap.set(cat.id, cat.name);
    });

    const groupedFeeds = new Map();
    feeds.forEach(feed => {
      const categoryName = feed.categoryId ? (categoryMap.get(feed.categoryId) || '') : '';
      if (!groupedFeeds.has(categoryName)) {
        groupedFeeds.set(categoryName, []);
      }
      groupedFeeds.get(categoryName).push(feed);
    });

    const outlines = [];

    const uncategorized = groupedFeeds.get('') || [];
    uncategorized.forEach(feed => {
      outlines.push({
        '@_title': feed.title || feed.url,
        '@_text': feed.title || feed.url,
        '@_type': 'rss',
        '@_xmlUrl': feed.url,
        '@_htmlUrl': feed.link || ''
      });
    });

    for (const [categoryName, categoryFeeds] of groupedFeeds) {
      if (categoryName === '') continue;
      
      const categoryOutlines = categoryFeeds.map(feed => ({
        '@_title': feed.title || feed.url,
        '@_text': feed.title || feed.url,
        '@_type': 'rss',
        '@_xmlUrl': feed.url,
        '@_htmlUrl': feed.link || ''
      }));

      outlines.push({
        '@_title': categoryName,
        '@_text': categoryName,
        'outline': categoryOutlines
      });
    }

    const opml = {
      '?xml': {
        '@_version': '1.0',
        '@_encoding': 'UTF-8'
      },
      opml: {
        '@_version': '2.0',
        head: {
          title: 'RSS Feeds Export',
          dateCreated: new Date().toUTCString()
        },
        body: {
          outline: outlines
        }
      }
    };

    return this.builder.build(opml);
  }

  importOpml(opmlContent) {
    const result = this.parser.parse(opmlContent);
    const feeds = [];

    if (!result?.opml?.body?.outline) {
      return feeds;
    }

    this.parseOutlines(result.opml.body.outline, feeds, null);
    return feeds;
  }

  parseOutlines(outlines, feeds, parentCategory) {
    if (!outlines) return;

    if (!Array.isArray(outlines)) {
      outlines = [outlines];
    }

    for (const outline of outlines) {
      const type = outline['@_type'] || '';
      const title = outline['@_title'] || outline['@_text'] || '';
      const xmlUrl = outline['@_xmlUrl'] || outline['@_url'] || '';
      const htmlUrl = outline['@_htmlUrl'] || '';

      if (type.toLowerCase() === 'rss' && xmlUrl) {
        feeds.push({
          title: title || xmlUrl,
          url: xmlUrl,
          link: htmlUrl,
          category: parentCategory
        });
      } else if (outline.outline) {
        this.parseOutlines(outline.outline, feeds, title || parentCategory);
      }
    }
  }
}

module.exports = OpmlHandler;
