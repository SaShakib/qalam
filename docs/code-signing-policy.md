# Code signing policy

Qalam's Windows downloads are signed so Windows can confirm they come from this project and haven't been changed. (Signing is being set up with SignPath Foundation; releases are signed from the first one after approval.)

**Free code signing provided by [SignPath.io](https://about.signpath.io/), certificate by [SignPath Foundation](https://signpath.org/).**

## What is signed

Only files built by GitHub Actions from this repository's source, by the [Release workflow](https://github.com/SaShakib/qalam/blob/main/.github/workflows/release.yml):

- `Qalam-Windows-x64.exe`, `Qalam-Windows-arm64.exe` (the app)
- `Qalam-Setup-x64.exe`, `Qalam-Setup-arm64.exe` (the installers)

Nothing built on a personal computer is signed with this certificate.

## Team and roles

| Role | Who |
|---|---|
| Committers and reviewers | [SaShakib](https://github.com/SaShakib) |
| Approvers (each signing request) | [SaShakib](https://github.com/SaShakib) |

Every release is approved by hand in SignPath before its files are signed.

## Privacy

Qalam doesn't collect or send any data about you or what you type. It works offline. The only network request it makes is to GitHub, to check for a new version of Qalam (and to download it, if you allow updates); GitHub's own [privacy statement](https://docs.github.com/site-policy/privacy-policies/github-general-privacy-statement) applies to that request.

The website (https://sashakib.github.io/qalam/) has no analytics or tracking. It loads fonts from Google Fonts and reads public release and question data from the GitHub API.
