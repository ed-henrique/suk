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

## Decisions

- Multiple cookies can connect to a single session, but they are not aware of
each other
- If all cookies to a session are used up, the reference to the session is lost

## Examples

- [Sample Server with Cookie Authentication](./examples/cookies/main.go)

## Roadmap

- [ ] Extensive testing
- [x] Make implementation concurrent-safe
- [ ] Use better algorithm for random and strong keys (refer to [this](https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go))
