# watHeaders - Quick analysis on the response headers provided by a list of hosts

> Warning: code is horrible as I have only used this for personal use. When I get the time I will make it less-horrible (and faster).

**It's the one thing pentesters hate doing...** collecting information regarding the _best practice_ response headers provided by a particular host, or list of hosts. 

The output file produced, is a CSV file containing the headers checked for, and the hosts that responded - with their results. Runtime flags are included so you can customize what represents success in the fields and what does not.

WatHeaders supports a configurable `hosts.json` file which allows you to add header `name` and `value` pairs to the array - supporting regex! So you can go through a list of hosts and determine if they respond with desired key:value pairings abiding by your supplied regex!

# Installation

```
go get -u github.com/glen-mac/watHeaders
```

# Usage

`$ watHeaders -i targets.txt`

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

```
$ cat wat.out

Host,Strict-Transport-Security,Public-Key-Pins,Cache-Control,X-Frame-Options,X-XSS-Protection,X-Content-Type-Options,Content-Security-Policy,Referrer-Policy
revoke.netgear.com,X,X,X,X,X,X,X,X
www.netgear.com,X,X, ,X,X,X,X,X
```

# Screenshot

<img src="https://i.imgur.com/FjYmMqX.png">

# To-Do

- Write better GoLang
- Optimize use of channels / passed structs
