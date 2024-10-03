# Contribution Guide 

There are many ways to be an open source contributor, and we're here to help you on your way! You may:

* Propose ideas in our 
  [discord](https://discord.gg/tbd)
* Raise an issue or feature request in our [issue tracker](https://github.com/TBD54566975/did-dht/issues)
* Help another contributor with one of their questions, or a code review
* Suggest improvements to our Getting Started documentation by supplying a Pull Request
* Evangelize our work together in conferences, podcasts, and social media spaces.

This guide is for you.

## 🎉 Hacktoberfest 2024 🎉

`did-dht` is a participating in Hacktoberfest 2024! We’re so excited for your contributions, and have created a wide variety of issues so that anyone can contribute. Whether you're a seasoned developer or a first-time open source contributor, there's something for everyone.

### Here's how you can get started:
1. Read the [code of conduct](https://github.com/taniashiba/did-dht/blob/main/CODE_OF_CONDUCT.md).
2. Choose a task from this project's Hacktoberfest issues in our [Project Hub](https://github.com/TBD54566975/did-dht/issues/292). Each issue has the 🏷️ `hacktoberfest` label.
5. Comment ".take" on the corresponding issue to get assigned the task.
6. Fork the repository and create a new branch for your work.
7. Make your changes and submit a pull request.
8. Wait for review and address any feedback.

### 🏆 Leaderboard & Prizes
Be among the top 10 with the most points to snag custom swag just for you from our TBD shop! To earn your place in the leaderboard, we have created a points system that is explained below. As you complete tasks, you will automatically be granted a certain # of points.

#### Point System
| Weight | Points Awarded | Description |
|---------|-------------|-------------|
| 🐭 **Small** | 5 points | For smaller tasks that take limited time to complete and/or don't require any product knowledge. |
| 🐰 **Medium** | 10 points | For average tasks that take additional time to complete and/or require some product knowledge. |
| 🐂 **Large** | 15 points | For heavy tasks that takes lots of time to complete and/or possibly require deep product knowledge. |

#### Prizes
Top 10 contributors with the most points will be awarded TBD x Hacktoberfest 2024 swag. The Top 3 contributors will have special swag customized with your GitHub handle in a very limited design. (more info in our Discord)



### 👩‍ Need help?
Need help or have questions? Feel free to reach out by connecting with us in our [Discord community](https://discord.gg/tbd) to get direct help from our team in the `#hacktoberfest` project channel.

Happy contributing!

---


## Development Prerequisites


| Requirement | Tested Version | Installation Instructions                             |
|-------------|----------------|-------------------------------------------------------|
| Go          | 1.23.2         | [go.dev](https://go.dev/doc/tutorial/compile-install) |
| Mage        | 1.15.0-5       | [magefile.org](https://magefile.org/)                   |

### Go

This project is written in Go, a modern, open source programming language. 

You may verify your `go` installation via the terminal:

```
$> go version
go version go1.23.2 darwin/amd64
```

If you do not have go, we recommend installing it by:

#### MacOS

##### Homebrew
```
$> brew install go
```

#### Linux

See the [Go Installation Documentation](https://go.dev/doc/install).

### Mage

The build is run by Mage.

You may verify your `mage` installation via the terminal:

```
$> mage --version
Mage Build Tool v1.15.0-5-g2385abb
Build Date: 2024-03-21T12:20:13-07:00
Commit: 2385abb
built with: go1.23.2
```

#### MacOS

##### Homebrew

```
$> brew install mage
```

#### Linux

Installation instructions are on the [Magefile home page](https://magefile.org/).
---

## Build (Mage)

```
$> mage build
```

## Test (Mage)

```
$> mage test
```

## Communications

### Issues

Anyone from the community is welcome (and encouraged!) to raise issues via 
[GitHub Issues](https://github.com/TBD54566975/did-dht/issues).
### Discussions

Design discussions and proposals take place in our [discord](https://discord.gg/tbd).

We advocate an asynchronous, written debate model - so write up your thoughts and invite the community to join in!

### Continuous Integration

Build and Test cycles are run on every commit to every branch on [GitHub Actions](https://github.com/TBD54566975/did-dht/actions).

## Contribution

We review contributions to the codebase via GitHub's Pull Request mechanism. We have 
the following guidelines to ease your experience and help our leads respond quickly 
to your valuable work:

* Start by proposing a change either in Issues (most appropriate for small 
  change requests or bug fixes) or in Discussions (most appropriate for design 
  and architecture considerations, proposing a new feature, or where you'd 
  like insight and feedback)
* Cultivate consensus around your ideas; the project leads will help you 
  pre-flight how beneficial the proposal might be to the project. Developing early 
  buy-in will help others understand what you're looking to do, and give you a 
  greater chance of your contributions making it into the codebase! No one wants to 
  see work done in an area that's unlikely to be incorporated into the codebase.
* Fork the repo into your own namespace/remote
* Work in a dedicated feature branch. Atlassian wrote a great 
  [description of this workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/feature-branch-workflow)
* When you're ready to offer your work to the project, first:
* Squash your commits into a single one (or an appropriate small number of commits), and 
  rebase atop the upstream `main` branch. This will limit the potential for merge 
  conflicts during review, and helps keep the audit trail clean. A good writeup for 
  how this is done is 
  [here](https://medium.com/@slamflipstrom/a-beginners-guide-to-squashing-commits-with-git-rebase-8185cf6e62ec), and if you're 
  having trouble - feel free to ask a member or the community for help or leave the commits as-is, and flag that you'd like 
  rebasing assistance in your PR! We're here to support you.
* Open a PR in the project to bring in the code from your feature branch.
* The maintainers noted in the `CODEOWNERS` file will review your PR and optionally 
  open a discussion about its contents before moving forward.
* Remain responsive to follow-up questions, be open to making requested changes, and...
  You're a contributor!
* And remember to respect everyone in our global development community. Guidelines 
  are established in our `CODE_OF_CONDUCT.md`.
