Asteroids are simple websites or personal wikis that can interact with the Realms on [gno.land](https://gno.land) blockchain. **You do not need** to be a technician to create an asteroid. However the installation requires to use `git` and `go` for now.

## Build the gnAsteroid executable

1. clone gnAsteroid: `git clone https://github.com/gno/gnAsteroid`
2. **Compile** with `go build -C cmd/ -o gnAsteroid`,
3. **Run** with `cmd/gnAsteroid -asteroid-dir=example`, 
4. **Visit** the example with a browser at http://localhost:8888.

## Create a GNO asteroid

Now our gnAsteroid command is compiled, we will create
an asteroid called "bob".

To create an asteroid "bob", we create a directory `bob`, 
with simply these 2 files inside:

**bob/index.md**
```md
Hello the Gnosmos! This is [Bob](about.md)'s blog :)

Take a look at my [gnoface](/r/demo/art/gnoface:1337).
```

**bob/about.md**
```md
I'm Bob from Neptune. In a previous life I was a sumo.
```

And that's it, we've created an asteroid with 2 pages.
Launch essentially like before:

`gnAsteroid -asteroid-name "Bob's asteroid" -asteroid-dir bob`

And pay a visit to your machine's local TCP port 8888, that is http://localhost:8888.

<img src=svg/colored-outlined/telescope.svg hspace=10 width=100 align=left />

**Quiz**: if you noticed, this asteroid can render not only
`about.md` and `index.md` but a third page, namely a realm 
`/r/demo/art/gnoface:1337` hosted and run by validators on the gno.land chain but 
shown on your asteroid. Can you guess what the `:1337` is and what it does? 

(**tip**: Using the asteroid, you can take a look at its source code on-chain. **tip 2**: If
you're too lazy to run your asteroid but want to [see it...](http://gno.land/r/demo/art/gnoface:1337)).

## Naming an asteroid

instead of specifying the name with `-asteroid-name <name>`, you may set it 
once and for all in a hidden file `.TITLE` at the root of your asteroid.

Just enter a one line name.

## Styling an asteroid

Asteroids are very rough rocks.

In the context of asteroids, we call **style** a set 
of 3 folders: `css`, `font` and `img`. 

<img src=svg/colored-outlined/darth-vader.svg hspace=10 width=100 align=left />

You may fork and modify the style in `default-style/`.  Then `gnAsteroid -style-dir=/path/to/your/style -asteroid-dir=example`

Asteroids css are not standardized yet. Don't hesitate to just fork and experiment for now (and contribute).  When no style is specified, a default style, embedded within the binary, is used.

## TODO

- [ ] realm to register, link and share asteroids (belts?)
- [ ] publishing
- [x] better `<table>` styling
- [ ] typography of the `<h1>-<h5>` titles (serif, sans?)
- [ ] converge gnoweb's views so this can remove ./views/
- [ ] ?help view for realms must use the value specified w/ -help-remote (and if not, -chainid must use -help-chainid)

## More 

More links may arrive soon, e.g.:

* gnAsteroid.style-gnosmos: a simple style you can use and modify.
* gnAsteroid.docker: a docker system to serve asteroids (especially on Akash)

## Credits

- [Gno.land](https://gno.land)'s gnoweb, the GNO chain, GNO VM, thanks to all the team!
- [Space Icons Set](https://iconduck.com/sets/space-icons-set) under the terms of Creative Commons Attribution 4, courtesy of [Agata Kuczmińska](https://iconduck.com/designers/agata-kuczminska). Amazing work she's done, thanks!
