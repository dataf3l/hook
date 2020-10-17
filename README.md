# hook: Continuous integration made simple.

This Simple tool can be used to automate processes based on URL requests.
For example, when you deploy a commit on gitlab, you can add a webhook notification
so that the process then proceeds to update some repo in a staging/dev server.

This is designed in order to save time for developers and guarantee all the code
that touches the production servers goes in via source control.


## How to use

In your repository, on local development environment you can add a file to your repo that looks like this:


```
{
    "commands": "./command.sh",
    "dev": "/some/folder",
    "master": "/some/other/folder/",
    "emails": [
        "yourself@gmail.com"
    ],
    "slack_webhook": "https://hooks.slack.com/services/...",
    "port":"1234",
     "smtp_from":"youremail@email.com",
  "smtp_host":"email.com",
  "smtp_port":"587",
  "smtp_user":"yourself@email.com",
  "smtp_pass":"yourpassword"
}
```

You will get an email notification and a slack notification when the tool is ran.
you can run the tool manually or automatically.
the commands in commands.sh can include popular ideas, like:

commands.sh:
```
git pull origin some-branch-name
go build
etc...
```

it is also recommended the tool is added to a systemctl unit file, so
you it is't stopped once the server restarts.

If any steps fail, the tool will stop running steps, this means 
you can for example, deploy only if the tool succeeded.

We hope this tool will help you save time.

Don't forget to star the repo! :)

