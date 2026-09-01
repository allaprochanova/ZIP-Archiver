// ZipArchiver.cs
// ZIP Archiver на C#

using System;
using System.Collections.Generic;
using System.IO;
using System.IO.Compression;
using System.Linq;

class ZipArchiver
{
    private readonly string archivePath;

    public ZipArchiver(string archivePath)
    {
        this.archivePath = archivePath;
    }

    public void Create(IEnumerable<string> files, int level)
    {
        using (var fs = new FileStream(archivePath, FileMode.Create))
        using (var zip = new ZipArchive(fs, ZipArchiveMode.Create))
        {
            foreach (var f in files)
            {
                AddToZip(zip, f, "");
            }
        }
        Console.WriteLine($"Архив {archivePath} создан.");
    }

    private void AddToZip(ZipArchive zip, string path, string basePath)
    {
        if (Directory.Exists(path))
        {
            foreach (var entry in Directory.GetFileSystemEntries(path))
            {
                string rel = basePath == "" ? Path.GetFileName(entry) : Path.Combine(basePath, Path.GetFileName(entry));
                AddToZip(zip, entry, rel);
            }
        }
        else
        {
            string entryName = basePath == "" ? Path.GetFileName(path) : basePath;
            zip.CreateEntryFromFile(path, entryName, CompressionLevel.Optimal);
        }
    }

    public void Extract(string outputDir)
    {
        using (var zip = ZipFile.OpenRead(archivePath))
        {
            foreach (var entry in zip.Entries)
            {
                string path = Path.Combine(outputDir, entry.FullName);
                if (entry.Name == "")
                {
                    Directory.CreateDirectory(path);
                }
                else
                {
                    Directory.CreateDirectory(Path.GetDirectoryName(path));
                    entry.ExtractToFile(path, true);
                }
            }
        }
        Console.WriteLine($"Архив {archivePath} распакован в {outputDir}");
    }

    public void List()
    {
        using (var zip = ZipFile.OpenRead(archivePath))
        {
            foreach (var entry in zip.Entries)
            {
                Console.WriteLine($"{entry.FullName}  {entry.Length} байт  {entry.LastWriteTime}");
            }
        }
    }

    public void Add(IEnumerable<string> files)
    {
        string tempPath = archivePath + ".tmp";
        // Копируем существующие записи
        using (var original = ZipFile.OpenRead(archivePath))
        using (var temp = new FileStream(tempPath, FileMode.Create))
        using (var zip = new ZipArchive(temp, ZipArchiveMode.Create))
        {
            foreach (var entry in original.Entries)
            {
                var newEntry = zip.CreateEntry(entry.FullName, entry.CompressionLevel);
                using (var src = entry.Open())
                using (var dst = newEntry.Open())
                {
                    src.CopyTo(dst);
                }
            }
            // Добавляем новые
            foreach (var f in files)
            {
                AddToZip(zip, f, "");
            }
        }
        File.Move(tempPath, archivePath, true);
        Console.WriteLine($"Файлы добавлены в {archivePath}");
    }

    public void Remove(IEnumerable<string> files)
    {
        string tempPath = archivePath + ".tmp";
        using (var original = ZipFile.OpenRead(archivePath))
        using (var temp = new FileStream(tempPath, FileMode.Create))
        using (var zip = new ZipArchive(temp, ZipArchiveMode.Create))
        {
            foreach (var entry in original.Entries)
            {
                if (files.Contains(entry.FullName))
                    continue;
                var newEntry = zip.CreateEntry(entry.FullName, entry.CompressionLevel);
                using (var src = entry.Open())
                using (var dst = newEntry.Open())
                {
                    src.CopyTo(dst);
                }
            }
        }
        File.Move(tempPath, archivePath, true);
        Console.WriteLine($"Файлы удалены из {archivePath}");
    }

    public void Comment(string text)
    {
        // В .NET нет встроенной поддержки комментариев, но можно добавить файл .comment
        Console.WriteLine("Комментарии не поддерживаются в стандартной библиотеке .NET.");
    }

    static void Main(string[] args)
    {
        if (args.Length == 0 || args[0] == "help")
        {
            Console.WriteLine(@"Использование: ZipArchiver <команда> [опции]
  create -a <archive> -f <files...> [-l level]
  extract -a <archive> [-o output]
  list -a <archive>
  add -a <archive> -f <files...>
  remove -a <archive> -f <files...>
  comment -a <archive> -t <text>");
            return;
        }

        string command = args[0];
        string archive = null;
        List<string> files = new List<string>();
        string output = ".";
        int level = 6;
        string text = null;

        for (int i = 1; i < args.Length; i++)
        {
            switch (args[i])
            {
                case "-a": archive = args[++i]; break;
                case "-f": files.AddRange(args[++i].Split(',')); break;
                case "-o": output = args[++i]; break;
                case "-l": level = int.Parse(args[++i]); break;
                case "-t": text = args[++i]; break;
            }
        }

        if (string.IsNullOrEmpty(archive))
        {
            Console.WriteLine("Ошибка: укажите -a");
            return;
        }

        var archiver = new ZipArchiver(archive);

        switch (command)
        {
            case "create":
                if (!files.Any()) { Console.WriteLine("Ошибка: укажите -f"); return; }
                archiver.Create(files, level);
                break;
            case "extract":
                archiver.Extract(output);
                break;
            case "list":
                archiver.List();
                break;
            case "add":
                if (!files.Any()) { Console.WriteLine("Ошибка: укажите -f"); return; }
                archiver.Add(files);
                break;
            case "remove":
                if (!files.Any()) { Console.WriteLine("Ошибка: укажите -f"); return; }
                archiver.Remove(files);
                break;
            case "comment":
                if (string.IsNullOrEmpty(text)) { Console.WriteLine("Ошибка: укажите -t"); return; }
                archiver.Comment(text);
                break;
            default:
                Console.WriteLine("Неизвестная команда");
                break;
        }
    }
}
