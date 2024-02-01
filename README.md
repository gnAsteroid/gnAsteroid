This helps creating **GNO asteroids**.

Asteroids are simple websites or personal wikis that can interact with the Realms on the [GNO blockchain](https://github.com/gnolang/gno). 

You can use it to setup personal blogs, organisation websites, research areas, presenting newly discovered gems in the gnoland or just to experiment. Like gno.land, it's kind of a smart-contract heightened wiki platform, but with redactional and stylistic autonomy.

gnAsteroid is a fork of the [gno.land](https://gno.land) website. However it doesn't intend to compete with it. Instead it's made to gravitate around it. 

# How to create a GNO asteroid

First, install gnAsteroid:

`git clone https://github.com/gno/gnAsteroid`

We will now create an asteroid for bob:

```
mkdir bob.blog/
echo "Hello the Gnosmos! This is [Bob](about)'s blog :)" > bob.blog/index.md
echo "I'm Bob from Neptune. In a previous life I was a sumo." > bob.blog/about.md
```

That's it, Bob created an asteroid with 2 pages.
Now Bob can organize pages as he wish, in directories and folders.

To launch a server:

```
make -C gnAsteroid/
gnAsteroid/website \
    -bind "127.0.0.1:2222" \
    -pages-dir bob.blog/wiki \
    -views-dir gnAsteroid/views
```

To test, he can visit http://127.0.0.1:2222

Tip: It is best to create a script for that.
See below how to create a script for automatic recompilation
and relaunch upon file changes.

# Watch files for auto-recompilation

## Linux

```bash
#!/bin/bash	
port=2222
while true; do
    make gnoweb2
    build/website2 \
        -bind "127.0.0.1:$port" \
        -pages-dir bob.blog/wiki \
        -views-dir gnAsteroid/views
    inotifywait -e modify,move,create,delete -r gnAsteroid/ -r bob.blog/
    pid=$(fuser -n tcp "$port" | cut -f2)
    if [[ -n $pid ]]; then
        kill -9 $pid
        sleep 0.2 
    fi
done
```

# Demo

