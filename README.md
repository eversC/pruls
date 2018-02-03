# pruls

### intro

(nope, I have no idea how to pronounce it either....it's just an ananym)

this is a go binary to simplify creation of a .tar.gz from a specific directory/file, and output to a specific google storage bucket.

### getting started

download [binary](https://github.com/eversC/pruls/releases)

* binary must exist on the host where the `targetdir` exists
* user running the binary must have:
    * read access to the `targetdir`
    * read access to the `accountKey` (usually your google .json key)
    * write access to the current work directory

it's then simply a case of setting env vars (see below) and running the binary, e.g. `"./pruls"`

it's recommended to stick to the `principle of least privilege` wrt required user access mentioned above. Create a new user specifically, and only give it the minimum access possible. The same applies to creating a google service account specifically for these backups.

### env vars

`appname`, `bucketname` and `targetdir` are all required

| name        |       required      | default      | description  |
| ------------- |:-------------:|:-------------:| :-----:|
| PRULS_ACCOUNTKEYABSPATH | n | /etc/secret-volume/auth.json| absolute path to account key (preferably json) |
| PRULS_APPNAME | y | | app name, used in the resulting filename |
| PRULS_BUCKETNAME | y | | name of google bucket |
| PRULS_FILEPREFIX | n | ""| a string to insert into the filename, could be useful for adding a version number |
| PRULS_TARGETDIRABSPATH | y | | absolute path of file/directory you want to back up |

### notes

* https://github.com/kelseyhightower/envconfig is used for env var config
* https://github.com/mholt/archiver for .tar.gz functionality
