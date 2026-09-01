// zip.ts
// ZIP Archiver на TypeScript

import * as fs from 'fs';
import * as path from 'path';
import AdmZip from 'adm-zip';

class ZipManager {
    private zip: AdmZip;

    constructor(private archivePath: string) {
        this.zip = fs.existsSync(archivePath) ? new AdmZip(archivePath) : new AdmZip();
    }

    private save(): void {
        this.zip.writeZip(this.archivePath);
    }

    create(files: string[], level: number = 6): void {
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
        this.save();
        console.log(`Архив ${this.archivePath} создан.`);
    }

    extract(outputDir: string = '.'): void {
        this.zip.extractAllTo(outputDir, true);
        console.log(`Архив ${this.archivePath} распакован в ${outputDir}`);
    }

    list(): void {
        const entries = this.zip.getEntries();
        for (const entry of entries) {
            console.log(`${entry.entryName}  ${entry.header.size} байт  ${entry.header.time}`);
        }
    }

    add(files: string[]): void {
        for (const f of files) {
            const stat = fs.statSync(f);
            if (stat.isDirectory()) {
                this.zip.addLocalFolder(f);
            } else {
                this.zip.addLocalFile(f);
            }
        }
        this.save();
        console.log(`Файлы добавлены в ${this.archivePath}`);
    }

    remove(files: string[]): void {
        for (const f of files) {
            this.zip.deleteEntry(f);
        }
        this.save();
        console.log(`Файлы удалены из ${this.archivePath}`);
    }

    comment(text: string): void {
        this.zip.setZipComment(text);
        this.save();
        console.log(`Комментарий добавлен в ${this.archivePath}`);
    }
}

function main(): void {
    const args = process.argv.slice(2);
    if (args.length === 0 || args[0] === 'help') {
        console.log(`Использование: ts-node zip.ts <команда> [опции]
  create    -a <archive> -f <files...> [-l level]
  extract   -a <archive> [-o output]
  list      -a <archive>
  add       -a <archive> -f <files...>
  remove    -a <archive> -f <files...>
  comment   -a <archive> -t <text>`);
        process.exit(0);
    }

    const command = args[0];
    const options: any = {};
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
