# Fuzzer for URL Parameters

## Usage 
`./fuzzup url regexp [wordlist]`

## Example
`./fuzzup 'https://example.com/page?{{}}={{}}&other=val' 'Requested Page:' wordlist.txt`

Fuzzup will read the wordlist and insert tab separated values into the url.
Lines matching the regexp are stripped from the response before it is hashed
and logged. Pages are reported to the user if they are the first page
with a unique hash.

For example, given the following wordlist:
```
page	test
page	test2
hello	world
```
the above command would do the following requests in order:
```
https://example.com/page?page=test&other=val
https://example.com/page?page=test2&other=val
https://example.com/page?hello=world&other=val
```

If the wordlist.txt is omitted, fuzzup will read from stdin.
