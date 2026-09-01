# zip.py
# ZIP Archiver на Python

import sys
import os
import zipfile
import argparse
from pathlib import Path
import shutil

class ZipManager:
    def __init__(self, archive_path):
        self.archive_path = archive_path

    def create(self, files, compression=zipfile.ZIP_DEFLATED, level=6):
        with zipfile.ZipFile(self.archive_path, 'w', compression=compression, compresslevel=level) as zf:
            for f in files:
                f = Path(f)
                if f.is_dir():
                    for root, dirs, files_in in os.walk(f):
                        for name in files_in:
                            file_path = Path(root) / name
                            arcname = file_path.relative_to(f.parent)
                            zf.write(file_path, arcname)
                else:
                    zf.write(f, f.name)
        print(f"Архив {self.archive_path} создан.")

    def extract(self, output_dir='.'):
        with zipfile.ZipFile(self.archive_path, 'r') as zf:
            zf.extractall(output_dir)
        print(f"Архив {self.archive_path} распакован в {output_dir}")

    def list(self):
        with zipfile.ZipFile(self.archive_path, 'r') as zf:
            for info in zf.infolist():
                print(f"{info.filename}  {info.file_size} байт  {info.date_time}")

    def add(self, files):
        with zipfile.ZipFile(self.archive_path, 'a') as zf:
            for f in files:
                f = Path(f)
                if f.is_dir():
                    for root, dirs, files_in in os.walk(f):
                        for name in files_in:
                            file_path = Path(root) / name
                            arcname = file_path.relative_to(f.parent)
                            zf.write(file_path, arcname)
                else:
                    zf.write(f, f.name)
        print(f"Файлы добавлены в {self.archive_path}")

    def remove(self, files):
        # Удаление требует пересоздания архива без указанных файлов
        temp = self.archive_path + ".tmp"
        with zipfile.ZipFile(self.archive_path, 'r') as zf_in:
            with zipfile.ZipFile(temp, 'w') as zf_out:
                for info in zf_in.infolist():
                    if info.filename not in files:
                        data = zf_in.read(info.filename)
                        zf_out.writestr(info, data)
        os.replace(temp, self.archive_path)
        print(f"Файлы удалены из {self.archive_path}")

    def comment(self, text):
        with zipfile.ZipFile(self.archive_path, 'a') as zf:
            zf.comment = text.encode('utf-8')
        print(f"Комментарий добавлен в {self.archive_path}")

def main():
    parser = argparse.ArgumentParser(description="ZIP Archiver (Python)")
    parser.add_argument("command", choices=["create", "extract", "list", "add", "remove", "comment", "help"],
                        help="Команда")
    parser.add_argument("-a", "--archive", help="Имя архива")
    parser.add_argument("-f", "--files", nargs="+", help="Файлы для добавления")
    parser.add_argument("-o", "--output", default=".", help="Папка для распаковки")
    parser.add_argument("-l", "--level", type=int, default=6, help="Уровень сжатия 0-9")
    parser.add_argument("-t", "--text", help="Текст комментария")
    args = parser.parse_args()

    if args.command == "help":
        print(__doc__)
        sys.exit(0)

    if not args.archive:
        print("Ошибка: укажите --archive")
        sys.exit(1)

    mgr = ZipManager(args.archive)

    if args.command == "create":
        if not args.files:
            print("Ошибка: укажите --files")
            sys.exit(1)
        mgr.create(args.files, level=args.level)
    elif args.command == "extract":
        mgr.extract(args.output)
    elif args.command == "list":
        mgr.list()
    elif args.command == "add":
        if not args.files:
            print("Ошибка: укажите --files")
            sys.exit(1)
        mgr.add(args.files)
    elif args.command == "remove":
        if not args.files:
            print("Ошибка: укажите --files")
            sys.exit(1)
        mgr.remove(args.files)
    elif args.command == "comment":
        if not args.text:
            print("Ошибка: укажите --text")
            sys.exit(1)
        mgr.comment(args.text)

if __name__ == "__main__":
    main()
