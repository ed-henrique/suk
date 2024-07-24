# SUCKS - Single Use Cookies (Killer Security)

What if instead of storing user info on the server side, in a JWT token, a token with a randomized session ID (with pretty good entropy) was used, expiring as soon as it was needed.

```mermaid
flowchart LR
    Q1(Is the user cookie valid?)
    P1(Generate new one/Invalidate old one)
    C1(Continue execution normally)
    C2(Authentication error)

    Q1-->|No|C2
    Q1-->|Yes|P1
    P1-->C1
```

## Decisions

- Multiple cookies can connect to a single session, but they are not aware of each other
- If all cookies to a session are used up, the reference to the session is lost

## Roadmap

- [ ] Make implementation concurrent-safe
- [ ] Use better algorithm for random and strong keys (refer to [this](https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go))
