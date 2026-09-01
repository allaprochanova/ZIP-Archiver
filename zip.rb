# zip.rb
# ZIP Archiver на Ruby

require 'zip'
require 'optparse'

class ZipManager
  def initialize(archive_path)
    @archive_path = archive_path
  end

  def create(files, level = 6)
    Zip::File.open(@archive_path, Zip::File::CREATE) do |zip|
      files.each do |f|
        add_to_zip(zip, f, '')
      end
    end
    puts "Архив #{@archive_path} создан."
  end

  def add_to_zip(zip, path, base)
    if File.directory?(path)
      Dir.entries(path).each do |entry|
        next if entry == '.' || entry == '..'
        full = File.join(path, entry)
        rel = base.empty? ? entry : File.join(base, entry)
        add_to_zip(zip, full, rel)
      end
    else
      zip.add(base.empty? ? File.basename(path) : base, path)
    end
  end

  def extract(output_dir = '.')
    Zip::File.open(@archive_path) do |zip|
      zip.each do |entry|
        dest = File.join(output_dir, entry.name)
        if entry.directory?
          FileUtils.mkdir_p(dest)
        else
          FileUtils.mkdir_p(File.dirname(dest))
          entry.extract(dest) { true }
        end
      end
    end
    puts "Архив #{@archive_path} распакован в #{output_dir}"
  end

  def list
    Zip::File.open(@archive_path) do |zip|
      zip.each do |entry|
        puts "#{entry.name}  #{entry.size} байт  #{entry.time}"
      end
    end
  end

  def add(files)
    # Пересоздаём архив
    temp = @archive_path + '.tmp'
    Zip::File.open(@archive_path) do |zip_in|
      Zip::File.open(temp, Zip::File::CREATE) do |zip_out|
        zip_in.each do |entry|
          zip_out.add(entry.name, entry.get_input_stream)
        end
        files.each do |f|
          add_to_zip(zip_out, f, '')
        end
      end
    end
    File.rename(temp, @archive_path)
    puts "Файлы добавлены в #{@archive_path}"
  end

  def remove(files)
    temp = @archive_path + '.tmp'
    Zip::File.open(@archive_path) do |zip_in|
      Zip::File.open(temp, Zip::File::CREATE) do |zip_out|
        zip_in.each do |entry|
          unless files.include?(entry.name)
            zip_out.add(entry.name, entry.get_input_stream)
          end
        end
      end
    end
    File.rename(temp, @archive_path)
    puts "Файлы удалены из #{@archive_path}"
  end

  def comment(text)
    # Комментарий не поддерживается в rubyzip, можно добавить файл
    puts "Комментарии не поддерживаются в rubyzip."
  end
end

def main
  options = {}
  OptionParser.new do |opts|
    opts.banner = "Использование: ruby zip.rb <команда> [опции]\n" +
                  "  create -a <archive> -f <files...> [-l level]\n" +
                  "  extract -a <archive> [-o output]\n" +
                  "  list -a <archive>\n" +
                  "  add -a <archive> -f <files...>\n" +
                  "  remove -a <archive> -f <files...>\n" +
                  "  comment -a <archive> -t <text>"
    opts.on('-a ARCHIVE') { |v| options[:archive] = v }
    opts.on('-f FILES') { |v| options[:files] = v.split(',') }
    opts.on('-o OUTPUT') { |v| options[:output] = v }
    opts.on('-l LEVEL') { |v| options[:level] = v.to_i }
    opts.on('-t TEXT') { |v| options[:text] = v }
  end.parse!

  if ARGV.empty?
    puts "Укажите команду"
    exit 1
  end

  command = ARGV[0]
  unless options[:archive]
    puts "Ошибка: укажите -a (архив)"
    exit 1
  end

  mgr = ZipManager.new(options[:archive])

  case command
  when 'create'
    unless options[:files]
      puts "Ошибка: укажите -f (файлы)"
      exit 1
    end
    level = options[:level] || 6
    mgr.create(options[:files], level)
  when 'extract'
    output = options[:output] || '.'
    mgr.extract(output)
  when 'list'
    mgr.list
  when 'add'
    unless options[:files]
      puts "Ошибка: укажите -f (файлы)"
      exit 1
    end
    mgr.add(options[:files])
  when 'remove'
    unless options[:files]
      puts "Ошибка: укажите -f (файлы)"
      exit 1
    end
    mgr.remove(options[:files])
  when 'comment'
    unless options[:text]
      puts "Ошибка: укажите -t (текст)"
      exit 1
    end
    mgr.comment(options[:text])
  else
    puts "Неизвестная команда"
    exit 1
  end
end

main if __FILE__ == $0
