package main

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
)

func main() {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}

	copyFolder := filepath.Join(cacheDir, "lcp-copy")

	flagCopy := flag.BoolP("copy", "c", false, "copy")
	flagPaste := flag.BoolP("paste", "p", false, "paste")
	flagClear := flag.BoolP("clear", "k", false, "clear")
	flagList := flag.BoolP("list", "l", false, "list")
	flagQuiet := flag.BoolP("quiet", "q", false, "quiet")

	flag.Parse()

	if *flagQuiet {
		log.SetOutput(io.Discard)
	}

	if *flagClear {
		os.RemoveAll(copyFolder)
		log.Println("cleared")
		os.Exit(0)
	}

	if *flagList {
		if _, err := os.Stat(copyFolder); os.IsNotExist(err) {
			log.Fatal("nothing to list")
		}

		var files []string
		err := filepath.Walk(copyFolder, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}
		for _, file := range files {
			println(file)
		}
		os.Exit(0)
	}

	if *flagCopy && *flagPaste {
		log.Fatal("copy and paste flags cannot be set at the same time")
	}

	if !(*flagCopy || *flagPaste) {
		log.Fatal("must set one of copy or paste flags")
	}

	if *flagCopy {
		if flag.NArg() == 0 {
			log.Fatal("nothing to copy provided")
		}

		if _, err := os.Stat(copyFolder); os.IsExist(err) {
			if err := os.RemoveAll(copyFolder); err != nil {
				log.Fatal(err)
			}
		}

		if _, err := os.Stat(copyFolder); os.IsNotExist(err) {
			if err := os.Mkdir(copyFolder, 0755); err != nil {
				log.Fatal(err)
			}
		}

		for q, file := range flag.Args() {
			if err := Copy(file, filepath.Join(copyFolder, filepath.Base(file))); err != nil {
				log.Fatal(err)
			}
			log.Println(q, "copied:", file)
		}
	}

	if *flagPaste {
		if _, err := os.Stat(copyFolder); os.IsNotExist(err) {
			log.Fatal("nothing to paste")
		}

		if Copy(copyFolder, wd) != nil {
			log.Fatal(err)
		}

		log.Println("pasted")
	}
}

func Copy(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return CopyDir(src, dst)
	} else {
		return CopyFile(src, dst)
	}
}

func CopyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	err = dstFile.Chmod(srcInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}

func CopyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dst, srcInfo.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		err = Copy(srcPath, dstPath)
		if err != nil {
			return err
		}
	}

	return nil
}
