class TrieNode {
  children: Map<string, TrieNode> = new Map();
  isEnd: boolean = false;
}

export class Trie {
  private root: TrieNode = new TrieNode();

  insert(word: string): void {
    let node = this.root;
    for (const char of word) {
      if (!node.children.has(char)) {
        node.children.set(char, new TrieNode());
      }
      node = node.children.get(char)!;
    }
    node.isEnd = true;
  }

  search(text: string): { word: string; start: number; end: number }[] {
    const results: { word: string; start: number; end: number }[] = [];
    const n = text.length;

    for (let i = 0; i < n; i++) {
      let node = this.root;
      let j = i;
      while (j < n && node.children.has(text[j])) {
        node = node.children.get(text[j])!;
        j++;
        if (node.isEnd) {
          results.push({
            word: text.substring(i, j),
            start: i,
            end: j
          });
        }
      }
    }

    return results;
  }
}
