# SUK - Single-use Keys

What if, instead of storing user information on the client side in a JWT token,
you used a token containing a randomized key that holds client information on
the server side? This session ID would be valid only for a specified duration
and would expire immediately after use. This approach enhances security and
minimizes the risk of unauthorized access.

This was intended to be used for web app authentication with HTTP cookies, but
other applications may find it useful as well.

```txt
Need authentication? ───────────────────────────────┐
├── Yes                                             │
│   └── Is the user key valid?                      │
│       ├── Yes                                     │
│       │   └── Generate new one/Invalidate old one │ 
│       │       └── Continue execution normally ────┘
│       └── No
│           └── Authentication error
└── No
    └── Well, ok then.
```


## Getting Started

### Getting SUK

```bash
go get -u github.com/ed-henrique/suk
```

### Running SUK

```go
package main

import (
    "github.com/ed-henrique/suk"
)

func main() {
    resource := "important stuff here!"

    ss, _ := suk.NewSessionStorage(suk.WithAutoClearExpiredKeys())

    // Sets resource to a randomly generated key
    key, _ := ss.Set(resource)

    // Gets the resource, invalidating the previous key
    resource, newKey, _ := ss.Get(key)

    // Removes both the key and the resource
    ss.Remove(newKey)
}
```

### Examples

- [Sample Server with Cookie Authentication](./examples/cookies/main.go)

### Documentation

Please refer to [this](https://pkg.go.dev/github.com/ed-henrique/suk).

## Decisions

- Multiple cookies can connect to a single session, but they are not aware of
each other
- If all cookies to a session are used up, the reference to the session is lost

## Roadmap

- [ ] Extensive testing
- [ ] Make implementation concurrent-safe (Currently, it is insanely difficult
to get a collision using the default 32 key length, but it is possible, so I'm
working to get around this chance)
- [ ] Use better algorithm for random and strong keys (refer to [this](https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go))
