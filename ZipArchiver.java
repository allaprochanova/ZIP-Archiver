// ZipArchiver.java
// ZIP Archiver на Java

import java.io.*;
import java.nio.file.*;
import java.util.*;
import java.util.zip.*;
import java.util.stream.*;

public class ZipArchiver {
    private final String archivePath;

    public ZipArchiver(String archivePath) {
        this.archivePath = archivePath;
    }

    public void create(List<String> files, int level) throws IOException {
        try (FileOutputStream fos = new FileOutputStream(archivePath);
             ZipOutputStream zos = new ZipOutputStream(fos)) {
            zos.setLevel(level);
            for (String f : files) {
                addToZip(zos, Paths.get(f), "");
            }
        }
        System.out.println("Архив " + archivePath + " создан.");
    }

    private void addToZip(ZipOutputStream zos, Path path, String base) throws IOException {
        if (Files.isDirectory(path)) {
            try (Stream<Path> paths = Files.list(path)) {
                for (Path p : paths.collect(Collectors.toList())) {
                    String rel = base.isEmpty() ? p.getFileName().toString() : base + "/" + p.getFileName().toString();
                    addToZip(zos, p, rel);
                }
            }
        } else {
            ZipEntry entry = new ZipEntry(base.isEmpty() ? path.getFileName().toString() : base);
            zos.putNextEntry(entry);
            Files.copy(path, zos);
            zos.closeEntry();
        }
    }

    public void extract(String outputDir) throws IOException {
        try (ZipInputStream zis = new ZipInputStream(new FileInputStream(archivePath))) {
            ZipEntry entry;
            while ((entry = zis.getNextEntry()) != null) {
                Path outPath = Paths.get(outputDir, entry.getName());
                if (entry.isDirectory()) {
                    Files.createDirectories(outPath);
                } else {
                    Files.createDirectories(outPath.getParent());
                    Files.copy(zis, outPath, StandardCopyOption.REPLACE_EXISTING);
                }
                zis.closeEntry();
            }
        }
        System.out.println("Архив " + archivePath + " распакован в " + outputDir);
    }

    public void list() throws IOException {
        try (ZipFile zipFile = new ZipFile(archivePath)) {
            Enumeration<? extends ZipEntry> entries = zipFile.entries();
            while (entries.hasMoreElements()) {
                ZipEntry entry = entries.nextElement();
                System.out.printf("%s  %d байт  %s%n", entry.getName(), entry.getSize(), entry.getTime());
            }
        }
    }

    public void add(List<String> files) throws IOException {
        // Пересоздаём архив с добавлением
        Path temp = Paths.get(archivePath + ".tmp");
        try (ZipOutputStream zos = new ZipOutputStream(Files.newOutputStream(temp))) {
            // Копируем существующие записи
            try (ZipFile zf = new ZipFile(archivePath)) {
                Enumeration<? extends ZipEntry> entries = zf.entries();
                while (entries.hasMoreElements()) {
                    ZipEntry entry = entries.nextElement();
                    zos.putNextEntry(new ZipEntry(entry.getName()));
                    try (InputStream is = zf.getInputStream(entry)) {
                        is.transferTo(zos);
                    }
                    zos.closeEntry();
                }
            }
            // Добавляем новые файлы
            for (String f : files) {
                addToZip(zos, Paths.get(f), "");
            }
        }
        Files.move(temp, Paths.get(archivePath), StandardCopyOption.REPLACE_EXISTING);
        System.out.println("Файлы добавлены в " + archivePath);
    }

    public void remove(List<String> files) throws IOException {
        Path temp = Paths.get(archivePath + ".tmp");
        try (ZipOutputStream zos = new ZipOutputStream(Files.newOutputStream(temp))) {
            try (ZipFile zf = new ZipFile(archivePath)) {
                Enumeration<? extends ZipEntry> entries = zf.entries();
                while (entries.hasMoreElements()) {
                    ZipEntry entry = entries.nextElement();
                    if (files.contains(entry.getName())) {
                        continue;
                    }
                    zos.putNextEntry(new ZipEntry(entry.getName()));
                    try (InputStream is = zf.getInputStream(entry)) {
                        is.transferTo(zos);
                    }
                    zos.closeEntry();
                }
            }
        }
        Files.move(temp, Paths.get(archivePath), StandardCopyOption.REPLACE_EXISTING);
        System.out.println("Файлы удалены из " + archivePath);
    }

    public void comment(String text) throws IOException {
        // Комментарий в Java ZIP не поддерживается (можно добавить как файл .comment)
        System.out.println("Комментарии не поддерживаются в стандартной библиотеке Java.");
    }

    public static void main(String[] args) throws IOException {
        if (args.length == 0 || args[0].equals("help")) {
            System.out.println("Использование: java ZipArchiver <команда> [опции]\n" +
                    "  create -a <archive> -f <files...> [-l level]\n" +
                    "  extract -a <archive> [-o output]\n" +
                    "  list -a <archive>\n" +
                    "  add -a <archive> -f <files...>\n" +
                    "  remove -a <archive> -f <files...>\n" +
                    "  comment -a <archive> -t <text>");
            System.exit(0);
        }

        String command = args[0];
        String archive = null;
        List<String> files = new ArrayList<>();
        String output = ".";
        int level = 6;
        String text = null;

        for (int i = 1; i < args.length; i++) {
            switch (args[i]) {
                case "-a":
                    archive = args[++i];
                    break;
                case "-f":
                    String[] parts = args[++i].split(",");
                    files.addAll(Arrays.asList(parts));
                    break;
                case "-o":
                    output = args[++i];
                    break;
                case "-l":
                    level = Integer.parseInt(args[++i]);
                    break;
                case "-t":
                    text = args[++i];
                    break;
            }
        }

        if (archive == null) {
            System.err.println("Ошибка: укажите -a");
            System.exit(1);
        }

        ZipArchiver archiver = new ZipArchiver(archive);

        switch (command) {
            case "create":
                if (files.isEmpty()) {
                    System.err.println("Ошибка: укажите -f");
                    System.exit(1);
                }
                archiver.create(files, level);
                break;
            case "extract":
                archiver.extract(output);
                break;
            case "list":
                archiver.list();
                break;
            case "add":
                if (files.isEmpty()) {
                    System.err.println("Ошибка: укажите -f");
                    System.exit(1);
                }
                archiver.add(files);
                break;
            case "remove":
                if (files.isEmpty()) {
                    System.err.println("Ошибка: укажите -f");
                    System.exit(1);
                }
                archiver.remove(files);
                break;
            case "comment":
                if (text == null) {
                    System.err.println("Ошибка: укажите -t");
                    System.exit(1);
                }
                archiver.comment(text);
                break;
            default:
                System.err.println("Неизвестная команда");
                System.exit(1);
        }
    }
}
