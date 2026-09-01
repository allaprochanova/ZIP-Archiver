<?php
// zip.php
// ZIP Archiver на PHP

if (php_sapi_name() !== 'cli') {
    die("Это консольное приложение.\n");
}

class ZipManager {
    private $archivePath;

    public function __construct($archivePath) {
        $this->archivePath = $archivePath;
    }

    public function create($files, $level = 6) {
        $zip = new ZipArchive();
        if ($zip->open($this->archivePath, ZipArchive::CREATE | ZipArchive::OVERWRITE) !== true) {
            throw new Exception("Не удалось создать архив");
        }
        foreach ($files as $f) {
            $this->addToZip($zip, $f, '');
        }
        $zip->close();
        echo "Архив {$this->archivePath} создан.\n";
    }

    private function addToZip($zip, $path, $base) {
        if (is_dir($path)) {
            $dir = opendir($path);
            while (($file = readdir($dir)) !== false) {
                if ($file == '.' || $file == '..') continue;
                $full = $path . DIRECTORY_SEPARATOR . $file;
                $rel = $base ? $base . '/' . $file : $file;
                $this->addToZip($zip, $full, $rel);
            }
            closedir($dir);
        } else {
            $zip->addFile($path, $base ?: basename($path));
        }
    }

    public function extract($outputDir = '.') {
        $zip = new ZipArchive();
        if ($zip->open($this->archivePath) !== true) {
            throw new Exception("Не удалось открыть архив");
        }
        $zip->extractTo($outputDir);
        $zip->close();
        echo "Архив {$this->archivePath} распакован в {$outputDir}\n";
    }

    public function list() {
        $zip = new ZipArchive();
        if ($zip->open($this->archivePath) !== true) {
            throw new Exception("Не удалось открыть архив");
        }
        for ($i = 0; $i < $zip->numFiles; $i++) {
            $stat = $zip->statIndex($i);
            echo $stat['name'] . "  " . $stat['size'] . " байт  " . date('Y-m-d H:i:s', $stat['mtime']) . "\n";
        }
        $zip->close();
    }

    public function add($files) {
        $zip = new ZipArchive();
        if ($zip->open($this->archivePath) !== true) {
            throw new Exception("Не удалось открыть архив");
        }
        foreach ($files as $f) {
            $this->addToZip($zip, $f, '');
        }
        $zip->close();
        echo "Файлы добавлены в {$this->archivePath}\n";
    }

    public function remove($files) {
        $zip = new ZipArchive();
        if ($zip->open($this->archivePath) !== true) {
            throw new Exception("Не удалось открыть архив");
        }
        foreach ($files as $f) {
            $zip->deleteName($f);
        }
        $zip->close();
        echo "Файлы удалены из {$this->archivePath}\n";
    }

    public function comment($text) {
        $zip = new ZipArchive();
        if ($zip->open($this->archivePath) !== true) {
            throw new Exception("Не удалось открыть архив");
        }
        $zip->setArchiveComment($text);
        $zip->close();
        echo "Комментарий добавлен в {$this->archivePath}\n";
    }
}

function main() {
    $args = array_slice($argv, 1);
    if (empty($args) || $args[0] == 'help') {
        echo "Использование: php zip.php <команда> [опции]\n";
        echo "  create -a <archive> -f <files...> [-l level]\n";
        echo "  extract -a <archive> [-o output]\n";
        echo "  list -a <archive>\n";
        echo "  add -a <archive> -f <files...>\n";
        echo "  remove -a <archive> -f <files...>\n";
        echo "  comment -a <archive> -t <text>\n";
        exit(0);
    }

    $command = $args[0];
    $options = [];
    for ($i = 1; $i < count($args); $i++) {
        if ($args[$i][0] == '-') {
            $key = substr($args[$i], 1);
            $value = $args[++$i] ?? '';
            $options[$key] = $value;
        }
    }

    if (!isset($options['a'])) {
        echo "Ошибка: укажите -a (архив)\n";
        exit(1);
    }

    $mgr = new ZipManager($options['a']);

    switch ($command) {
        case 'create':
            if (!isset($options['f'])) {
                echo "Ошибка: укажите -f (файлы)\n";
                exit(1);
            }
            $files = explode(',', $options['f']);
            $level = isset($options['l']) ? (int)$options['l'] : 6;
            $mgr->create($files, $level);
            break;
        case 'extract':
            $output = $options['o'] ?? '.';
            $mgr->extract($output);
            break;
        case 'list':
            $mgr->list();
            break;
        case 'add':
            if (!isset($options['f'])) {
                echo "Ошибка: укажите -f (файлы)\n";
                exit(1);
            }
            $mgr->add(explode(',', $options['f']));
            break;
        case 'remove':
            if (!isset($options['f'])) {
                echo "Ошибка: укажите -f (файлы)\n";
                exit(1);
            }
            $mgr->remove(explode(',', $options['f']));
            break;
        case 'comment':
            if (!isset($options['t'])) {
                echo "Ошибка: укажите -t (текст)\n";
                exit(1);
            }
            $mgr->comment($options['t']);
            break;
        default:
            echo "Неизвестная команда\n";
            exit(1);
    }
}

main();
