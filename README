# Fuzzer for URL Parameters

## Usage 
`./fuzzup 'https://example.com/page?{{}}={{}}&other=val' wordlist.txt`

Fuzzup will read the wordlist and insert tab separated values into the url.
The response is hashed and logged, any changes in the hashed page are reported
to the user, with the url that triggered the change.

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
