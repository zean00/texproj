# texproj
Project text into pixels (image) and later can be use for perceptual hashing  


###Usage

```
go get github.com/zean00/texproj

texproj -i article.txt -d idwiki.2d.norm.bin -o article.png -r 512

-i input_text
-d dictionary_file
-o output_image
-r resolution (square)

```

###Notes

Dictionary is generated from Indonesian wikipedia that have been processed with lexvec algorithm (similar to word2vec) with dimension reduction (tSNE). The quality of dictionary (coordinate distribution) should be improved by more training iteration (currently it's only 150 iteration) that requires high resource computing (~90GB of RAM for 38,000 vocabulary)  

Dictionary format 

```
dan 0.5748094991825784 0.48004685403813
yang 0.5752662199006712 0.4808744672747991
di 0.5760991390440556 0.4813181129173777

word x_coordinate y_coordinate
```