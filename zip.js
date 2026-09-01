// zip.js
// ZIP Archiver на JavaScript (Node.js)

const fs = require('fs');
const path = require('path');
const AdmZip = require('adm-zip');

class ZipManager {
    constructor(archivePath) {
        this.archivePath = archivePath;
        this.zip = null;
        this._load();
    }

    _load() {
        if (fs.existsSync(this.archivePath)) {
            this.zip = new AdmZip(this.archivePath);
        } else {
            this.zip = new AdmZip();
        }
    }

    _save() {
        this.zip.writeZip(this.archivePath);
    }

    create(files, level = 6) {
        this.zip = new AdmZip();
        for (const f of files) {
            const stat = fs.statSync(f);
            if (stat.isDirectory()) {
                this.zip.addLocalFolder(f);
            } else {
                this.zip.addLocalFile(f);
            }
        }
        this.zip.setDefaultCompression(level === 0 ? 'STORE' : 'DEFLATE');
        this._save();
        console.log(`Архив ${this.archivePath} создан.`);
    }

    extract(outputDir = '.') {
        if (!this.zip) throw new Error('Архив не загружен');
        this.zip.extractAllTo(outputDir, true);
        console.log(`Архив ${this.archivePath} распакован в ${outputDir}`);
    }

    list() {
        if (!this.zip) throw new Error('Архив не загружен');
        const entries = this.zip.getEntries();
        for (const entry of entries) {
            console.log(`${entry.entryName}  ${entry.header.size} байт  ${entry.header.time}`);
        }
    }

    add(files) {
        if (!this.zip) throw new Error('Архив не загружен');
        for (const f of files) {
            const stat = fs.statSync(f);
            if (stat.isDirectory()) {
                this.zip.addLocalFolder(f);
            } else {
                this.zip.addLocalFile(f);
            }
        }
        this._save();
        console.log(`Файлы добавлены в ${this.archivePath}`);
    }

    remove(files) {
        if (!this.zip) throw new Error('Архив не загружен');
        for (const f of files) {
            this.zip.deleteEntry(f);
        }
        this._save();
        console.log(`Файлы удалены из ${this.archivePath}`);
    }

    comment(text) {
        if (!this.zip) throw new Error('Архив не загружен');
        this.zip.setZipComment(text);
        this._save();
        console.log(`Комментарий добавлен в ${this.archivePath}`);
    }
}

function main() {
    const args = process.argv.slice(2);
    if (args.length === 0 || args[0] === 'help') {
        console.log(`Использование: node zip.js <команда> [опции]
  create    -a <archive> -f <files...> [-l level]
  extract   -a <archive> [-o output]
  list      -a <archive>
  add       -a <archive> -f <files...>
  remove    -a <archive> -f <files...>
  comment   -a <archive> -t <text>`);
        process.exit(0);
    }

    const command = args[0];
    const options = {};
    for (let i = 1; i < args.length; i++) {
        if (args[i].startsWith('-')) {
            const key = args[i].replace(/^-+/, '');
            const value = args[++i];
            options[key] = value;
        }
    }

    if (!options.a) {
        console.error('Ошибка: укажите -a (архив)');
        process.exit(1);
    }

    const mgr = new ZipManager(options.a);

    switch (command) {
        case 'create':
            if (!options.f) {
                console.error('Ошибка: укажите -f (файлы)');
                process.exit(1);
            }
            const files = options.f.split(',');
            const level = options.l ? parseInt(options.l) : 6;
            mgr.create(files, level);
            break;
        case 'extract':
            mgr.extract(options.o || '.');
            break;
        case 'list':
            mgr.list();
            break;
        case 'add':
            if (!options.f) {
                console.error('Ошибка: укажите -f (файлы)');
                process.exit(1);
            }
            mgr.add(options.f.split(','));
            break;
        case 'remove':
            if (!options.f) {
                console.error('Ошибка: укажите -f (файлы)');
                process.exit(1);
            }
            mgr.remove(options.f.split(','));
            break;
        case 'comment':
            if (!options.t) {
                console.error('Ошибка: укажите -t (текст)');
                process.exit(1);
            }
            mgr.comment(options.t);
            break;
        default:
            console.error('Неизвестная команда');
            process.exit(1);
    }
}

main();
