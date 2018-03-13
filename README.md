# watHeaders - Quick analysis on the response headers provided by a list of hosts

> Warning: code is horrible as I have only used this for personal use. When I get the time I will make it less-horrible (and faster).

It's the one thing pentesters hate doing.. collecting information regarding the _best practice_ response headers provided by a particular host, or list of hosts. 

Why do it by hand when you can automate it? And in Go! 

WatHeaders supports a configurable `hosts.json` file which allows you to add header `name` and `value` pairs to the array - supporting regex! So you can go through a list of hosts and determine if they respond with desired key:value pairings abiding by your supplied regex!

# Installation

```
go get -u github.com/glen-mac/watHeaders
```

# Usage

`watHeaders -i subdomains.txt -o wat.out`

```
Usage of ./watHeaders:
  -case-sensitive
        Case-sensitive string matching
  -f string
        Output 'found' marking (default " ")
  -i string
        Newline separated hosts file
  -l string
        File containing headers (default "headers.json")
  -m string
        Output 'missing' marking (default "X")
  -o string
        Output file (default "wat.out")
  -r uint
        Timeout for connections (default 1)
  -t int
        Number of concurrent threads (default 10)
```

# Screenshot

<img src="https://i.imgur.com/wEab0I3.jpg">

# To-Do

- Write better GoLang
- Optimize use of channels / passed structs
