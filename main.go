package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	clr "image/color"
	"image/png"

	"github.com/dchest/blake2b"
	ltrie "github.com/zean00/LevenshteinTrie"
)

type point struct {
	x float64
	y float64
}

type texel struct {
	word  string
	count int
	seq   int
	coord point
	color []byte
}

type texels []texel

func (slice texels) Len() int {
	return len(slice)
}

func (slice texels) Less(i, j int) bool {
	return slice[i].count < slice[j].count
}

func (slice texels) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

var trie *ltrie.TrieNode

func main() {
	flset := flag.NewFlagSet("", flag.ExitOnError)
	iOpt := flset.String("i", "", "input text")
	dOpt := flset.String("d", "", "dictionary path")
	oOpt := flset.String("o", "out.png", "output file")
	rOpt := flset.Int("r", 256, "output resolution")
	flset.Parse(os.Args[1:])

	if *iOpt == "" {
		log.Fatal("Missing input file")
	}
	if *dOpt == "" {
		log.Fatal("Missing dictionary file")
	}

	loadTVector(*dOpt)
	process(*iOpt, *oOpt, *rOpt)
}

func loadTVector(path string) {
	//path := "./idwiki.2d.norm.bin"

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	trie = ltrie.NewTrie()
	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		fmt.Fprintf(os.Stderr, fmt.Sprintf("Scanning line %d\r", i))
		str := strings.Split(scanner.Text(), " ")
		x, err := strconv.ParseFloat(str[1], 64)
		if err != nil {
			continue
		}
		y, err := strconv.ParseFloat(str[2], 64)
		if err != nil {
			continue
		}
		p := &point{
			x: x,
			y: y,
		}
		trie.Add(str[0], p)
		i++
	}

	fmt.Println("Load dictionary : ", i)

}

func process(path, out string, res int) string {
	words := loadArticle(path)
	texs := []texel{}
	i := 0
	for _, w := range words {
		if w == " " || w == "\n" || w == "" {
			continue
		}
		texs = append(texs, toTexel(w, i))
		i++
	}
	log.Println("Word length ", len(words))
	texs = countTexel(texs)
	log.Println("Sorted texels ", len(texs))
	return paintImage(texs, out, res)

}

func loadArticle(path string) []string {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Article not found")
	}
	txt := string(dat)
	txt = strings.Replace(txt, "-", " ", -1)
	txt = strings.Replace(txt, ".", " ", -1)
	txt = strings.Replace(txt, ",", " ", -1)
	txt = strings.Replace(txt, "\n", " ", -1)
	return strings.Split(txt, " ")
}

func cleanWord(word string) string {
	word = strings.Trim(word, " ")
	word = strings.Replace(word, "\"", "", -1)
	word = strings.Replace(word, "'", "", -1)
	word = strings.Replace(word, ",", "", -1)
	word = strings.Replace(word, ".", "", -1)
	word = strings.Replace(word, "!", "", -1)
	word = strings.Replace(word, "?", "", -1)
	return strings.ToLower(word)
}

func toTexel(word string, seq int) texel {
	word = cleanWord(word)
	te := texel{
		word:  word,
		seq:   seq,
		count: 1,
	}

	return te
}

func countTexel(texs []texel) []texel {
	words := make(map[string]texel)
	for _, t := range texs {
		t.seq = (t.seq + words[t.word].seq) / 2
		t.count += words[t.word].count
		words[t.word] = t
	}
	txls := texels{}
	unknown := 0
	//i := 0
	for _, t := range words {
		node := trie.Get(t.word)
		if node == nil {
			qs := trie.Levenshtein(t.word, 3)
			if len(qs) == 0 {
				unknown++
				continue
			}
			node = qs[len(qs)-1].Node
		}

		t.coord = *node.GetInfo().(*point)
		t.color = hash(t.word, 3)
		txls = append(txls, t)
		//i++
	}

	sort.Sort(sort.Reverse(txls))
	log.Println("Unknown words ", unknown)
	return txls
}

func paintImage(data []texel, filename string, dim int) string {
	myimage := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{dim, dim}})
	for _, t := range data {
		c := clr.RGBA{uint8(t.color[0]), uint8(t.color[1]), uint8(t.color[2]), 255}
		px := t.count / 4
		x := int(t.coord.x * float64(dim))
		y := int(t.coord.y * float64(dim))
		drawArea(myimage, x, y, px, t.seq, c)
		//drawArea(myimage, x, y, px, 0, c)
	}
	myfile, _ := os.Create(filename)
	png.Encode(myfile, myimage)
	return filename
}

func drawArea(img *image.RGBA, x, y, pixel, shift int, c clr.RGBA) {
	x += int(math.Sqrt(float64(shift)))
	y += int(math.Sqrt(float64(shift)))
	//log.Println("Drawing ", x, " ", y, " color ", c)
	for m := (x - pixel); m < (x + pixel); m++ {
		for n := (y - pixel); n < (y + pixel); n++ {
			img.Set(m, n, c)
		}
	}
}

func hash(s string, size int) []byte {
	h, _ := blake2b.New(&blake2b.Config{Size: uint8(size)})
	h.Write([]byte(s))
	sum := h.Sum(nil)
	return sum
}
