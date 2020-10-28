# imgcat

Tool to output images in the terminal. Built with [bubbletea](https://github.com/charmbracelet/bubbletea)

## install

### homebrew

```
brew install trashhalo/homebrew-brews/imgcat
```

### prebuilt packages

Prebuilt packages can be found at the [releases page](https://github.com/trashhalo/imgcat/releases)

### aur

```
yay -S imgcat
```

## sample output
```
imgcat https://i.redd.it/65fmdbh1ja951.jpg
```

![sample](./sample.png)

## files on disk

```
imgcat *.jpg
```

- j, down: next image
- k, up: previous image

## gifs

imgcat supports gifs and yes they will animate on your terminal. The quality of that animation seems largely dependant on your terminal. [Kitty](https://sw.kovidgoyal.net/kitty/#) runs them best. 
