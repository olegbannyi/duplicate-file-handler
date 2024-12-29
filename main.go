package main

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type App struct {
	dir           string
	fileExtension string
	order         int
	files         map[int64][]string
}

func main() {
	app := NewApp()
	app.Init()
	app.Scan()
	app.Print()
	app.HandleDuplicates()
}

func NewApp() *App {
	return &App{
		files: make(map[int64][]string),
	}
}

func (a *App) Init() {
	if len(os.Args) < 2 {
		a.OnError(fmt.Errorf("Directory is not specified"))
	}

	dir := os.Args[1]
	// Check if the directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		a.OnError(fmt.Errorf("Directory does not exist"))
	}

	a.dir = dir

	fmt.Println("Enter file format:")
	reader := bufio.NewReader(os.Stdin)
	fileExtension, err := reader.ReadString('\n')
	if err != nil {
		a.OnError(err)
	}
	fileExtension = strings.TrimSpace(fileExtension)
	fileExtension = strings.Trim(fileExtension, ".")

	if len(fileExtension) == 0 {
		a.fileExtension = fileExtension
	} else {
		a.fileExtension = "." + fileExtension
	}

	fmt.Println("Size sorting options:")
	fmt.Println("1. Descending")
	fmt.Println("2. Ascending")

	for {
		fmt.Println("Enter a sorting option:")
		var order int

		fmt.Scan(&order)

		if order == 1 || order == 2 {
			a.order = order
			break
		} else {
			fmt.Println("Wrong option")
		}
	}
}

func (a *App) Scan() {
	anyFile := len(a.fileExtension) == 0
	// Get all the files in the directory
	err := filepath.Walk(a.dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if anyFile || filepath.Ext(path) == a.fileExtension {
			a.files[info.Size()] = append(a.files[info.Size()], path)
		}

		return nil
	})

	if err != nil {
		a.OnError(err)
	}
}

func (a *App) Print() {
	sizes := a.FilesSizes()

	for _, size := range sizes {
		fmt.Println(size, "bytes")
		files := a.files[size]
		for _, file := range files {
			fmt.Println(file)
		}
		fmt.Println()
	}
}

func (a *App) FilesSizes() []int64 {
	sizes := make([]int64, 0, len(a.files))

	for size := range a.files {
		sizes = append(sizes, size)
	}

	sort.Slice(sizes, func(i, j int) bool {
		if a.order == 1 {
			return sizes[i] > sizes[j]
		} else {
			return sizes[i] < sizes[j]
		}
	})

	return sizes
}

func (a *App) DuplictedFilesSizes(list map[int64]map[string][]string) []int64 {
	sizes := make([]int64, 0, len(list))

	for size := range list {
		sizes = append(sizes, size)
	}

	sort.Slice(sizes, func(i, j int) bool {
		if a.order == 1 {
			return sizes[i] > sizes[j]
		} else {
			return sizes[i] < sizes[j]
		}
	})

	return sizes
}

func (a *App) HandleDuplicates() {
	if !a.CheckDuplicates() {
		return
	}
	duplicates := a.GetDuplicates()
	max := a.PrintDuplicates(duplicates)

	if max != 0 {
		a.DoRemoveDuplicates(max, duplicates)
	}
}

func (a *App) CheckDuplicates() bool {
	fmt.Println("Check for duplicates?")
	reader := bufio.NewReader(os.Stdin)

	for {
		answer, err := reader.ReadString('\n')
		if err != nil {
			a.OnError(err)
		}
		answer = strings.TrimSpace(answer)

		if answer == "yes" {
			return true
		} else if answer == "no" {
			return false
		} else {
			fmt.Println("Wrong option")
		}
	}
}

func (a *App) RemoveDuplicates(max int) []int {
	fmt.Println("Delete files?")
	reader := bufio.NewReader(os.Stdin)

	for {
		answer, err := reader.ReadString('\n')
		if err != nil {
			a.OnError(err)
		}
		answer = strings.TrimSpace(answer)

		if answer == "yes" {
			return a.FileNumbers(max)
		} else if answer == "no" {
			return []int{}
		} else {
			fmt.Println("Wrong option")
		}
	}
}

func (a *App) DoRemoveDuplicates(max int, duplicates map[int64]map[string][]string) {
	index := 0
	freedSpace := int64(0)
	if conditions := a.RemoveDuplicates(max); len(conditions) > 0 {
		sizes := a.DuplictedFilesSizes(duplicates)
		for _, size := range sizes {
			filesMap := duplicates[size]
			for _, files := range filesMap {
				for _, file := range files {
					index++
					if a.Contains(conditions, index) {
						os.Remove(file)
						freedSpace += size
					}
				}
			}
		}
		fmt.Printf("Total freed up space: %d bytes\n", freedSpace)
	}
}

func (a *App) FileNumbers(max int) []int {
	var result []int
outer:
	for {
		fmt.Println("Enter file numbers to delete:")
		reader := bufio.NewReader(os.Stdin)
		answer, err := reader.ReadString('\n')
		if err != nil {
			a.OnError(err)
		}
		answer = strings.TrimSpace(answer)

		numbers := strings.Split(answer, " ")
		result = make([]int, 0)
		for _, number := range numbers {
			if n, err := strconv.Atoi(number); err != nil {
				fmt.Println("Wrong format")
				continue outer
			} else if n < 1 || n > max {
				fmt.Println("Wrong format")
				continue outer
			} else {
				result = append(result, n)
			}
		}

		break
	}

	return result
}

func (a *App) PrintDuplicates(duplicates map[int64]map[string][]string) int {
	index := 0
	sizes := a.DuplictedFilesSizes(duplicates)

	for _, size := range sizes {
		fmt.Println(size, "bytes")
		filesMap := duplicates[size]
		for hash, files := range filesMap {
			fmt.Println("Hash:", hash)
			for _, file := range files {
				index++
				fmt.Printf("%d. %s\n", index, file)
			}
		}
		fmt.Println()
	}

	return index
}

func (a *App) GetDuplicates() map[int64]map[string][]string {
	duplicates := make(map[int64]map[string][]string, len(a.files))

	for size, files := range a.files {
		if len(files) > 1 {
			for _, file := range files {
				if _, ok := duplicates[size]; !ok {
					duplicates[size] = make(map[string][]string)
				}
				hash := a.FileHash(file)
				duplicates[size][hash] = append(duplicates[size][hash], file)
			}
		}
	}

	for size, files := range duplicates {
		for hash, files := range files {
			if len(files) == 1 {
				delete(duplicates[size], hash)
			}
		}
	}

	return duplicates
}

func (a *App) FileHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		a.OnError(err)
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		a.OnError(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

func (a *App) OnError(err error) {
	fmt.Println(err)
	os.Exit(0)
}

func (a *App) Contains(list []int, value int) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}

	return false
}
