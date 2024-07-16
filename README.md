Asteroids are simple websites or personal wikis that can interact with the Realms on the [GNO blockchain](https://github.com/gnolang/gno). 

You can use it to setup personal blogs, organisation websites, research areas, presenting newly discovered gems in the gnoland or just to experiment. Like gno.land, it's a kind of wiki-like platform that can link to smart-contract on Gno, with redactional and stylistic autonomy.

gnAsteroid is built on the [gnoweb package](https://github.com/gnolang/gno/tree/master/gno.land/pkg/gnoweb) from [gno.land](https://gno.land). It doesn't intend to compete with it of course, since it's plugged onto the same realms (smart contracts) but asteroids are made to gravitate around gno.land using the same idea.

# Minimal example

* **Compile** using `go build -o gnAsteroid`,
* **Run** with `./gnAsteroid -asteroid-dir=example`, 
* **Visit** with a browser.

# To create a GNO asteroid

1. install gnAsteroid: `git clone https://github.com/gno/gnAsteroid`
2. Create an asteroid, let's say we're bob:

```
mkdir -p ~asteroids/bob/
cd ~/asteroids
echo "Hello the Gnosmos! This is [Bob](about)'s blog :)" > bob/index.md
echo "I'm Bob from Neptune. In a previous life I was a sumo." > bob/about.md
```

That's it, an asteroid with 2 pages.
Launch essentially like before:
```bash
gnAsteroid \
    -bind 127.0.0.1:8888 \
    -asteroid-name "Bob's asteroid" \
    -asteroid-dir ~/asteroids/bob
```
Test by visiting [http://127.0.0.1:8888](http://127.0.0.1:8888).

# Styling an asteroid?

Asteroids are very rough rocks.

In the context of asteroids, we call **style** a set 
of 3 folders: `css`, `font` and `img`. You may fork and
modify the style at [grepsuzette/gnAsteroid.style-gnosmos](https://github.com/grepsuzette/gnAsteroid.style-gnosmos).

Then `gnAsteroid -style-dir=/path/to/your/style -asteroid-dir=example`

When no style is specified, a style similar to gnosmos, embedded 
within the binary, is used.

# TODO

- The doc could be an asteroid.
- a realm to register asteroid belts?
- querying asteroid belts to show links?

# More 

This is just the server.

More links will arrive soon:

* gnAsteroid.style-gnosmos: a simple style you can use and modify.
* gnAsteroid.docker: a docker system to serve asteroids (especially on Akash)

