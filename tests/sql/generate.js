// Create languages
var sqls = [];
for (var i = 1; i <= 10000; i++) {
  const id = i.toString().padStart(5, '0');
  sqls.push(`('Language ${id}')`);
}
console.log(`INSERT INTO languages (name) VALUES \n${sqls.join(',\n')};`);

// Create tags
var sqls = [];
for (var i = 1; i <= 10000; i++) {
  const id = i.toString().padStart(5, '0');
  sqls.push(`('Tag ${id}')`);
}
console.log(`INSERT INTO tags (name) VALUES \n${sqls.join(',\n')};`);

// Create categories
var sqls = [];
for (var i = 1; i <= 10000; i++) {
  const id = i.toString().padStart(5, '0');
  sqls.push(`('Category ${id}', 'category-${id}', 10000)`);
}
console.log(`INSERT INTO categories (name, slug, language_id) VALUES \n${sqls.join(',\n')};`);

// Create categories - tags relations
var sqls = [];
for (var i = 1; i <= 10000; i++) {
  sqls.push(`(10000, ${i})`);
}
console.log(`INSERT INTO categories_tags (categories, tags) VALUES \n${sqls.join(',\n')};`);

// Create blogs
var sqls = [];
for (var i = 1; i <= 10000; i++) {
  const id = i.toString().padStart(5, '0');
  sqls.push(`('Blog ${id}', 10000)`);
}
console.log(`INSERT INTO blogs (name, category_id) VALUES \n${sqls.join(',\n')};`);
