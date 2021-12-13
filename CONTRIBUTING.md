# Table of Contents
- [Table of Contents](#table-of-contents)
- [Contents](#contents)
  - [Contributing with code to c13n-go](#contributing-with-code-to-c13n-go)
    - [We Develop on Github](#we-develop-on-github)
    - [We Use Github Flow model](#we-use-github-flow-model)
    - [Make good commits](#make-good-commits)
  - [Reporting Bugs](#reporting-bugs)
    - [Report bugs using Github's issues](#report-bugs-using-githubs-issues)
    - [Write bug reports with detail, background, and sample code](#write-bug-reports-with-detail-background-and-sample-code)
  - [License](#license)

# Contents

## Contributing with code to c13n-go

Any input by the community is welcome, so we want to make the process of contributing as easy as possible. Your contributions may include bug reports, discussions on current state of project, proposals for new features or fix submissions.

### We Develop on Github
We use github to host code, track issues, and review Pull Requests. The project's life cycle is taking place here.

### We Use [Github Flow](https://docs.github.com/en/get-started/quickstart/github-flow) model
We actively review and discuss pull requests.

The desired format & flow for this procedure is as follows:

1. Fork the repo under your namespace. Then clone locally:
```
git clone github.com/<my_namespace>/c13n-go
```
2. Create a branch from `develop`.
```
git checkout develop
git checkout -b <my_branch_name>
```
3. Add your changes, by creating simple & understandable commits. Do not include artifacts, media or binary files to your commits.
```
git add <file_1>, <file_2>, ...
git commit -m "Descriptive commit message"
git push origin <my_branch_name>
```
4. Make the Pull Request. Don't forget to write a small description of the changes you are introducing. If the Pull Request related to a reported issue, refer to it by issue number.
5. Stay active for the review phase, as fixes and/or tweaks may be required for your PR to be approved.
6. Final Merge from the dev team.

### Make good commits
These are the seven rules of a great Git commit messages:

* Separate subject from body with a blank line
* Limit the subject line to 50 characters (if possible)
* Capitalize the subject line
* Do not end the subject line with a period
* Use the imperative mood in the subject line
* Wrap the body at 72 characters
* Use the body to explain what and why vs. how

Try to solve a single problem per commit.
If your description ends up too long, that’s an indication that you probably need to split up your commit.

Create commit or pull-request descriptions that are self-contained.
This benefits both the maintainers and reviewers.

## Reporting Bugs

### Report bugs using Github's issues
We use GitHub issues to track public bugs. Report a bug by opening a new [issue](https://github.com/c13n-io/c13n-go/issues).

### Write bug reports with detail, background, and sample code

**Great** Bug Reports tend to have:

- A quick summary and/or background
- Steps to reproduce
  - Be specific
  - Give sample code, if possible and applicable.
- What you expected would happen
- What actually happens
- Notes (including why you think this might be happening, or possible solutions you tried that didn't work)

## License

This project is licensed under the [MIT License](http://choosealicense.com/licenses/mit/).

When you submit code changes, your submissions are understood to be under the same [MIT License](http://choosealicense.com/licenses/mit/) that covers the project. Feel free to contact the maintainers if that's a concern.
