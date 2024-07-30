Welcome to gnAsteroid.

Asteroids are simple websites or personal wikis that can interact with the Realms on [gno.land](https://gno.land) blockchain. 

## Build the gnAsteroid executable

Note: *You do not need* to be a technician to create an asteroid. However the first step still requires to use `git` and `go`. If you manage to build the gnAsteroid executable below, the rest is going to be easy.

<img src=svg/colored-outlined/space-rover-2.svg width=100 align=left style="padding-right: 25px;" />

1. clone gnAsteroid: `git clone https://github.com/gno/gnAsteroid`
2. **Compile** with `go build -C cmd/ -o gnAsteroid`,
3. **Run** with `cmd/gnAsteroid -asteroid-dir=example`, 
4. **Visit** the example with a browser at http://localhost:8888.

## Creating an asteroid

We built a `gnAsteroid` (or `gnAsteroid.exe` on Windows) file in the previous step. We will now create
an asteroid. Our first fictional asteroid will be about someone called "bob".

To set up "bob" as an asteroid, you'll need to
 create a directory named bob containing two
 simple files:

```plaintext
Directory Structure:
bob/
├── index.md
└── about.md
```

To create an asteroid "bob", we create a directory `bob`, 
with these two simple files inside:

**bob/index.md**
```md
Hello the Gnosmos! This is [Bob](about.md)'s blog :)

Take a look at my [gnoface](/r/demo/art/gnoface:1337).
```

**bob/about.md**
```md
I'm Bob from Neptune. In a previous life, I was a sumo.
```

The format is markdown. If you are not familiar with it, it's very easy to learn.
A button is like `[text](link)` (producing [text](link) ← this won't work but you may click). 
Links used here are all relative: they point to this asteroid.

And that's it, we've created an asteroid with 2 pages. It's finished. 
Launch essentially like before:

`gnAsteroid -asteroid-name "Bob's asteroid" -asteroid-dir bob`

And pay a visit to your machine's local TCP port 8888, that is http://localhost:8888.

<img src=svg/colored-outlined/telescope.svg hspace=10 width=100 align=left />

**Quiz**: if you noticed, this asteroid can render not only
`about.md` and `index.md` but a third page, namely a realm 
`/r/demo/art/gnoface:1337` hosted and run by validators on the gno.land chain but 
shown on your asteroid. Can you guess what the `:1337` is and what it does? 

(**tip**: Using the asteroid, you can take a look at its source code on-chain. **tip 2**: you can also [see it on gno.land...](http://gno.land/r/demo/art/gnoface:1337) ← click here and then locate the source button, it should be somewhere on the top-right of the page.)

## Naming an asteroid

instead of specifying the name with `-asteroid-name <name>`, you may set it 
once in a hidden file `.TITLE` at the root of your asteroid. For example, *Precious... My Precious*.

## Styling an asteroid

Asteroids are very rough rocks.

In the context of asteroids, we call **style** a set 
of 3 folders: `css`, `font` and `img`. 

<style type="text/css">
img#really:hover { 
    content:url("svg/colored-outlined/chewbacca.svg"); 
}
</style>
<img id=really title="Damn, Han Solo!" src=svg/colored-outlined/darth-vader.svg hspace=10 width=100 align=left />

You may fork and modify the style in `default-style/`.  Then launch something like `gnAsteroid.exe -style-dir=/path/to/your/style -asteroid-dir=example`

Asteroids css are not standardized yet. Don't hesitate to just fork and experiment for now (and contribute).  When no style is specified, a default style, embedded within the executable, is used.

## Publishing

Publishing your asteroid means to share it with other people. 

This is a work in progress, but [here's what works](publishing/) for now.

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
* [gnAsteroid.docker](https://github.com/grepsuzette/gnAsteroid.docker): a Dockerfile to help serve asteroids (especially on [Akash](https://console.akash.network))

## Credits

- [Gno.land](https://gno.land)'s gnoweb, the GNO chain, GNO VM, thanks to all the team!
- [Space Icons Set](https://iconduck.com/sets/space-icons-set) under the terms of Creative Commons Attribution 4, courtesy of [Agata Kuczmińska](https://iconduck.com/designers/agata-kuczminska). Amazing work, thanks!
- This work is published under the same terms as GNO itself. ☮ 
