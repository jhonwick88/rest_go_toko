import sys

path = r'H:\FlutterProject\rest_go_toko\database\database.go'
with open(path, 'r', encoding='utf-8') as f:
    content = f.read()

target = '	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {'
replacement = '	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000; PRAGMA foreign_keys = ON;"); err != nil {'

content = content.replace(target, replacement)

with open(path, 'w', encoding='utf-8') as f:
    f.write(content)
