const http = require('http');

function post(path, data) {
  return new Promise((resolve, reject) => {
    const json = JSON.stringify(data);
    const req = http.request({
      hostname: 'localhost', port: 8080, path, method: 'POST',
      headers: { 'Content-Type': 'application/json; charset=utf-8', 'Content-Length': Buffer.byteLength(json) }
    }, res => {
      let body = '';
      res.on('data', c => body += c);
      res.on('end', () => {
        try { resolve(JSON.parse(body)); } catch(e) { resolve({ raw: body }); }
      });
    });
    req.on('error', reject);
    req.write(json);
    req.end();
  });
}

async function seed() {
  const cats = [{ name: '技术' }, { name: '生活' }, { name: '读书' }];
  for (const c of cats) {
    try { const r = await post('/api/categories', c); console.log('Cat:', r.name || r.raw); } catch(e) { console.log('Cat skip'); }
  }

  const articles = [
    { title: '理解 Go 调度器：从协程到 Goroutine', content: '## GMP 模型\n\nGo 的调度器基于 GMP 模型：G（Goroutine）是用户态轻量线程，M（Machine）是 OS 线程，P（Processor）是逻辑处理器。\n\nP 的数量由 GOMAXPROCS 决定。理解 GMP 是写出高效 Go 程序的关键。', category: '技术', tags: 'Go,并发,后端', is_published: true },
    { title: 'Python 协程 vs Go 协程', content: '## 两种协程模型\n\nPython asyncio 基于事件循环，单线程。Go goroutine 基于 GMP 调度器，多线程。\n\n选择建议：Web 爬虫用 Python，API 服务和数据处理用 Go。', category: '技术', tags: 'Python,Go,后端', is_published: true },
    { title: 'SQLite 全文搜索实战笔记', content: '## FTS5\n\nSQLite FTS5 是内置全文搜索引擎，比 LIKE 快几十倍。支持布尔搜索、前缀匹配和结果排序。\n\n中文分词需要额外处理，简单场景下布尔搜索已够用。', category: '技术', tags: 'SQLite,后端,架构', is_published: true },
    { title: 'Docker 容器化个人项目', content: '## 多阶段构建\n\n即使个人项目，Docker 也能带来环境一致性、快速部署和依赖隔离的好处。\n\n使用多阶段构建，最终镜像不到 20MB，启动秒级。', category: '技术', tags: 'Docker,架构,后端', is_published: true },
    { title: '为什么你应该写博客', content: 'Paul Graham 说过：写作不只是记录想法，写作本身就是思考的过程。\n\n写下来才知道自己是不是真的懂了。不需要完美的第一篇。Done is better than perfect。', category: '生活', tags: '读书,设计', is_published: true },
    { title: '秋日徒步：西湖到灵隐寺', content: '从断桥出发，沿北山路走到灵隐寺，约 5 公里。\n\n深秋的杭州最美，梧桐叶铺了一地金黄。到了灵隐寺，先闻到炒栗子的味道。寺里银杏全黄了，阳光从叶子缝隙漏下来。很值。', category: '生活', tags: '徒步,咖啡', is_published: true },
    { title: '手冲咖啡入门：从磨豆到注水', content: '四个变量：研磨度像细砂糖、水温 88-93度、粉水比 1:15、萃取时间 2:30-3:00。\n\n最重要的事：豆子要好。新鲜烘焙的豆子和超市货架上的豆子，差别远大于技术和机器的差异。', category: '生活', tags: '咖啡', is_published: true },
    { title: '黑客与画家 读书笔记', content: '黑客和画家最像：都是创作者。财富不是零和游戏，可以通过创造价值来获得。\n\n编程语言影响思考：更强大的语言让你思考更高层次的问题。这是 Paul Graham 最重要的洞见之一。', category: '读书', tags: '读书,设计', is_published: true },
    { title: '构建个人知识管理系统', content: '收集、整理、输出、复习。\n\n工具链：Obsidian 笔记、Todoist 任务、Readwise 阅读、博客写作。每周花 30 分钟整理，每月回顾一次。知识管理不在于工具，在于习惯。', category: '技术', tags: '设计,架构', is_published: true },
    { title: '成为更好的开发者：2024 年回顾', content: '最大的收获不是某个具体技术，而是一种心态：持续做小事。\n\n每天写一点代码，读一点书，写几个字。一年下来回头看，发现自己走了很远。\n\n明年目标：50 篇博客、学 Rust、做一个开源项目。', category: '生活', tags: '读书,生活', is_published: true },
  ];

  for (const a of articles) {
    try { const r = await post('/api/articles', a); console.log('Article:', r.title); } catch(e) { console.log('Article err'); }
  }

  // Notes
  const notes = [
    '今天用 Go 写了个小工具，自动把 Markdown 渲染成 HTML，配合 goldmark 库很顺手 🎉',
    '秋日午后阳光正好，泡了杯手冲，埃塞俄比亚 耶加雪菲，柑橘调很明显',
    '读完了《黑客与画家》，Paul Graham 对程序员创造力的理解太深刻了。推荐给每个写代码的人',
    '终于把博客的 3D 标签云调好了，Sphere 算法比想象中简单。数学真美',
    '今天走了 10 公里，从龙井村走到九溪。秋天的杭州，每一步都是风景',
  ];

  for (const content of notes) {
    try { const r = await post('/api/notes', { content, is_published: true }); console.log('Note:', r.id); } catch(e) { console.log('Note err'); }
  }

  console.log('\n=== Seed Complete ===');
}

seed().catch(e => console.error(e));
