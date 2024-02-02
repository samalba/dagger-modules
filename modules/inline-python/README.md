# inline-python-mod

Dagger module for inlining Python code in a pipeline

## Examples

### Get time and time on my local timezone

```sh
dagger call -m github.com/samalba/inline-python-mod@main \
    code --code "import datetime as dt; print(dt.datetime.now())"
```

### Get my public IP

```sh
dagger call -m github.com/samalba/inline-python-mod@main \
    with-package --name "requests" \
    code --code "import sys,requests as r; sys.stdout.write(r.get('https://api.ipify.org?format=json').json()['ip'])"
```
