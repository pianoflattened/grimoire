# grimoire

supplies local file usercripts to firefox (or any compatible fork) through 
h43z's [beval][beval] extension. any script it has access to will be "hot
reloaded" (? if this is correct usage of the buzzword?) on a page refresh, so 
once you get this set up you shouldn't have to think about what's going on 
with it at all and you can Just Program

## usage
this only works with firefox! first you want to install the [extension][ext].
once you have that installed, you can download or clone the repo anywhere you
like, [`go build`][build] to get your executable, and set up the included manifest file 
`beval.json` according to [this][mfstdoc] set of instructions--there is a 
section under "setup" titled "App manifest" under which you will find specific 
instructions for windows, mac, and linux. mozilla also has a useful 
[README][mozrdme] that explains how to set up an example native messaging app. 
at some point i will just distribute a binary and write some cute little 
installer scripts and then you  won't have to think so hard about it. the main 
detail to remember is that this will only work if the path to the executable 
`grimoire` on your computer is the same as in the manifest. unless you plan on 
downloading this repo to `/path/to/grimoire`, you will have to change it!

when you open up your browser for the first time, if you look in the directory
containing the executable, you will find it has spit out `defaults.toml`--
```toml
# if you're seeing this, you need to write a config file!!
# put a file named `conf.toml' in the same directory as your executable and
# fill it out as the example is done here and then "uncomment" (remove #s from) 
# the lines to get this program sending scripts to firefox

refresh = 100

# [[spell]]
# pattern = "*.example.com/*"
# dir = "sites/example"
```
the program does not like to read from this config--it is a backup that it 
produces on failure to find your config. as the comments explain, you should
create a `conf.toml` to put right next to your executable. this is where you
will explain to it what scripts should be used on what websites

the main thing you will be doing in `conf.toml` is adding "spells". you tell it
"on `pattern` kind of url, i want to load the files in `dir` folder". if you
don't know how toml works, each new entry has to have `[[spell]]` as a header.
`refresh` is an optional setting that controls the amount of time that the 
program will wait before responding to a second edit right after a first one, 
in milliseconds. if you set it to 0, your program will usually see 2 or 3 quick 
writes to any newly saved file, as that is usually just what a text editor 
does--so you do want to have Some number there. but i wouldn't worry too much 
about it

## !!!! WARNING !!!!
this program works as a drop-in replacement for h43z's native messaging client,
which creates a unix pipe and directs all input and output of the extensions
context through that pipe. this uses its dumb (only meaning that it takes any 
string recieved from its native app and evals it without thinking) browser-side 
background script and uses it as an injection point for locally hosted 
userscripts. taking away the funny unix pipe SHOULD make this work easily enough 
for windows and mac. i am not remotely well-studied in this kind of topic but
this kind of setup seems irredeemably bad from a security standpoint. so if you
are going to use it i would take care only to do it on a personal machine that
you share with no one. i believe the license this software comes with already
absolves me of it, but i will reiterate that part of it: i am not responsible
for whatever malicious scripts you might accidentally give access to your web
browser using this tool

[beval]: https://github.com/h43z/beval
[ext]: https://addons.mozilla.org/en-US/firefox/addon/beval/
[build]: https://go.dev/doc/tutorial/compile-install
[mfstdoc]: https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions/Native_messaging#setup
[mozrdme]: https://github.com/SphinxKnight/webextensions-examples/tree/master/native-messaging
