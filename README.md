# JeffCI
A custom CI bot made for the Go on z/OS team

## Usage

Currently, a new 'build & test & package' run will be triggered by the creation of a new pull request, and any subsequent pushes to open pull requests.

### Re-Try
A build can be retried with the following command:
```
@JeffCI RETRY
```
_Or any case-insensitive variation of "respin, rebuild, retest" with a dash (re-try) or without_

### Example
An example command exists to trigger a sample task that has no actual effect on the status of the PR:
```
@JeffCI example
```
This can be used to test that the bot is working. It should also be used as a reference point for adding new tasks/runs to the bot.

## Setup

1. Ensure that the config in `main.go` is set appropriately. The AppID will stay common to the 'JeffCI' GHE app, but the InstallationID will change from org-to-org. Also ensure that the bots private key exists in the bots base directory. A new one can always be generated from within the GHE apps settings page.

2. Add `cronjob.sh` to the host machines crontab. This script just checks to see if the bot is already running. If it isn't it pulls any new changes from the repo, re-builds the bot, and starts a new session.

> [!NOTE] The bot currently runs on `zosgo@csz25086.pok.stglabs.ibm.com` (linux)
