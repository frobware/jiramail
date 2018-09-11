# jiramail

The `jiramail` is mail transport for Atlassian's Jira service written in Go. This utility
stores data (`projects`, `sprints`, `issues`, `comments`, etc.) from jira to local maildir.
Optionally for making changes, it provides an SMTP interface.

# Configuration

Example:

```yaml
core:
  loglevel: debug
  logfile: ~/logs/jiramail.log
  lockdir: ~/tmp/jiramail.lock
smtp:
  addr: 127.0.0.1:10025
remote:
  coreos:
    destdir: ~/Mail/jira/coreos
    baseurl: https://jira.coreos.com
    username: legionus
    password: PRIVATE
    projectmatch: DEVEXP
    boardmatch: DEVEXEP
    delete: remove
```

# SMTP server

The utility provides a special SMTP server so that you can make changes in the Jira. It does not
send emails anywhere, but only performs API requests based on them. So this is not a real SMTP server.

The server understands the context and the necessary parameters when you reply to the generated message.
So it matters to which message you reply.

If you reply to the message and send it to:

### reply@jira

* reply to `issue` or `comment` will add new comment;
* reply to `project` will create a new issue.

### edit@jira

The reply to this address is used to edit the `Subject` and the body of the message. Be careful and do not
forget about quoting in your mail client. This operation is valid for `issues` and `comments`.

### bot@jira

This is a special address where the body of the message is treated as a sequence of commands.
This operation is valid for `issues`

#### Change labels

```
labels add "label one" "label two"
labels remove no-qe
```

#### Change state

```
state "in progress"
state To Do
```
Argument is not case sensitive.

#### Change priority

```
priority high
priority low
```
Argument is not case sensitive.

#### Assignee

```
assignee to me
assignee to legionus
```