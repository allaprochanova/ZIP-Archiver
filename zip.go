// zip.go
// ZIP Archiver на Go

package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type ZipManager struct {
	archivePath string
}

func NewZipManager(archivePath string) *ZipManager {
	return &ZipManager{archivePath: archivePath}
}

func (z *ZipManager) create(files []string, level int) error {
	// Создаём архив
	out, err := os.Create(z.archivePath)
	if err != nil {
		return err
	}
	defer out.Close()

	zipWriter := zip.NewWriter(out)
	defer zipWriter.Close()

	// Устанавливаем уровень сжатия (в archive/zip нет прямого управления уровнем, используем Store или Deflate)
	// По умолчанию Deflate, можно только выбрать метод Store (0) или Deflate (1)
	// Но уровень сжатия не регулируется в стандартной библиотеке Go.
	// Можно использовать внешнюю библиотеку или игнорировать.

	for _, f := range files {
		err := addToZip(zipWriter, f, "")
		if err != nil {
			return err
		}
	}
	fmt.Printf("Архив %s создан.\n", z.archivePath)
	return nil
}

func addToZip(zw *zip.Writer, src, base string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		entries, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			child := filepath.Join(src, entry.Name())
			rel := filepath.Join(base, entry.Name())
			if err := addToZip(zw, child, rel); err != nil {
				return err
			}
		}
		return nil
	}
	// Файл
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	// Определяем имя в архиве
	arcName := base
	if arcName == "" {
		arcName = filepath.Base(src)
	}
	header := &zip.FileHeader{
		Name:   arcName,
		Method: zip.Deflate,
	}
	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, file)
	return err
}

func (z *ZipManager) extract(outputDir string) error {
	reader, err := zip.OpenReader(z.archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		path := filepath.Join(outputDir, file.Name)
		if file.FileInfo().IsDir() {
			os.MkdirAll(path, os.ModePerm)
			continue
		}
		os.MkdirAll(filepath.Dir(path), os.ModePerm)
		rc, err := file.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		outFile, err := os.Create(path)
		if err != nil {
			return err
		}
		defer outFile.Close()
		_, err = io.Copy(outFile, rc)
		if err != nil {
			return err
		}
	}
	fmt.Printf("Архив %s распакован в %s\n", z.archivePath, outputDir)
	return nil
}

func (z *ZipManager) list() error {
	reader, err := zip.OpenReader(z.archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	for _, file := range reader.File {
		fmt.Printf("%s  %d байт  %s\n", file.Name, file.UncompressedSize64, file.Modified)
	}
	return nil
}

func (z *ZipManager) add(files []string) error {
	// Для добавления нужно пересоздать архив
	tmpPath := z.archivePath + ".tmp"
	// Копируем существующий архив во временный
	err := copyFile(z.archivePath, tmpPath)
	if err != nil {
		return err
	}
	// Открываем существующий для чтения
	reader, err := zip.OpenReader(z.archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Создаём новый архив
	out, err := os.Create(z.archivePath)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	defer zw.Close()

	// Копируем существующие записи
	for _, f := range reader.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		header := f.FileHeader
		w, err := zw.CreateHeader(&header)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(w, rc)
		rc.Close()
		if err != nil {
			return err
		}
	}
	// Добавляем новые файлы
	for _, f := range files {
		err := addToZip(zw, f, "")
		if err != nil {
			return err
		}
	}
	fmt.Printf("Файлы добавлены в %s\n", z.archivePath)
	return nil
}

func (z *ZipManager) remove(files []string) error {
	// Удаление через пересоздание архива без указанных файлов
	tmpPath := z.archivePath + ".tmp"
	reader, err := zip.OpenReader(z.archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	out, err := os.Create(z.archivePath)
	if err != nil {
		return err
	}
	defer out.Close()
	zw := zip.NewWriter(out)
	defer zw.Close()

	for _, f := range reader.File {
		skip := false
		for _, name := range files {
			if f.Name == name {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		header := f.FileHeader
		w, err := zw.CreateHeader(&header)
		if err != nil {
			rc.Close()
			return err
		}
		_, err = io.Copy(w, rc)
		rc.Close()
		if err != nil {
			return err
		}
	}
	fmt.Printf("Файлы удалены из %s\n", z.archivePath)
	return nil
}

func (z *ZipManager) comment(text string) error {
	// Комментарий в ZIP хранится в конце, но archive/zip не поддерживает запись комментария в стандартной библиотеке.
	// Пропустим для Go.
	fmt.Println("Комментарии не поддерживаются в стандартной библиотеке Go.")
	return nil
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()
	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()
	_, err = io.Copy(dest, source)
	return err
}

func main() {
	var archive, output, text string
	var level int
	var files string
	var cmd string

	flag.StringVar(&cmd, "cmd", "list", "Команда: create, extract, list, add, remove, comment")
	flag.StringVar(&archive, "a", "", "Архив")
	flag.StringVar(&files, "f", "", "Файлы (через запятую)")
	flag.StringVar(&output, "o", ".", "Папка для распаковки")
	flag.IntVar(&level, "l", 6, "Уровень сжатия")
	flag.StringVar(&text, "t", "", "Текст комментария")
	flag.Usage = func() {
		fmt.Println("Использование: go run zip.go -cmd <команда> -a <архив> [опции]")
	}
	flag.Parse()

	if archive == "" {
		fmt.Println("Ошибка: укажите -a")
		os.Exit(1)
	}

	mgr := NewZipManager(archive)

	switch cmd {
	case "create":
		if files == "" {
			fmt.Println("Ошибка: укажите -f")
			os.Exit(1)
		}
		err := mgr.create(strings.Split(files, ","), level)
		if err != nil {
			fmt.Println("Ошибка:", err)
			os.Exit(1)
		}
	case "extract":
		err := mgr.extract(output)
		if err != nil {
			fmt.Println("Ошибка:", err)
			os.Exit(1)
		}
	case "list":
		err := mgr.list()
		if err != nil {
			fmt.Println("Ошибка:", err)
			os.Exit(1)
		}
	case "add":
		if files == "" {
			fmt.Println("Ошибка: укажите -f")
			os.Exit(1)
		}
		err := mgr.add(strings.Split(files, ","))
		if err != nil {
			fmt.Println("Ошибка:", err)
			os.Exit(1)
		}
	case "remove":
		if files == "" {
			fmt.Println("Ошибка: укажите -f")
			os.Exit(1)
		}
		err := mgr.remove(strings.Split(files, ","))
		if err != nil {
			fmt.Println("Ошибка:", err)
			os.Exit(1)
		}
	case "comment":
		if text == "" {
			fmt.Println("Ошибка: укажите -t")
			os.Exit(1)
		}
		err := mgr.comment(text)
		if err != nil {
			fmt.Println("Ошибка:", err)
			os.Exit(1)
		}
	default:
		fmt.Println("Неизвестная команда")
	}
}
