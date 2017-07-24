# speakers

This is a site designed to list and recruit speakers for various types of events. It was originally written for speakers for tech events, but there aren't any design limitations on its use.

## Language
It is written in Go and SCSS/SASS. Go was chosen because it makes deployment easier for non technical users as there are no dependencies with the installed application, all you need is the css, js and templates folder installed along with the compiled app. SCSS/SASS was chosen as it makes CSS easier to produce for those not very familiar with CSS.

## Config file
The config file, speakers.cfg, contains 10 parameters:
- The Port is the port the the app listens on. It is a numeric value.
- The Name of the app. This is displayed on most pages as a title.
- Initials is a string which appears at the start of every page title. It makes it easier for your users to know which tabs apply to the speakers app.
- Domain is the url of the website. It is used for links in activation codes and password reset codes.
- Secure is whether the site is http or https, true indicates https
- ActivateEMail is the email address the activate e-mail is from, eg activate@…
- SendMessageEMail is the email address messages from other users are from, eg message@…
- NewPassEMail is the email address the password reset e-mail is from, eg passwordreset@…
- MailGunDomain is the sending domain for MailGun.
- MailKey is the api key for the Mailgun sending domain.

## Encryption and security
There is a fair chance you will get hacked. Speakers has been built assuming this will be the case. All private information is either encrypted or hashed. Scrypt was chosen for hashing as it, at the moment, seems pretty secure and was designed such that using a GPU to attack the hashing won't be as helpful as it is with some other hashing algorithms. Replace <E Mail Salt> with the hash vale the e-mails 

AES is the encryption standard. The variant of AES — AES 128, AES 192 or AES 256 — depends on the length of the encryption key used when building the app. 16 characters for AES 128, 24 for AES 192 and 32 for AES 256. The key is not stored in the repository or on disk as that would make an attackers life a bit easier. To set a key and encryption level substitute the <encryption key> in the build statement with your desired key.

## Compiling
To build the program to run on the host machine the command is:

go build -ldflags "-X main.key=encryption key|E Mail Salt"

To build for linux:

GOOS=linux GOARCH=amd64 go build -ldflags "-X main.key=encryption key|E Mail Salt"

To build for a Raspberry Pi:

GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-X main.key=encryption key|E Mail Salt"

## Database
MongoDB was chosen for the database. It was picked because only four tables are needed for the app to run. It isn't subject to SQL injection, or needs an ORM, it has replication and good read times.

The users table has details of the users, the locations where they are prepared to speak and a list of their talks

The tags table is a list of tags for talk types, devops, languages etc.

The places table has a list of speaking locations, usually cities as most speakers do it for free, so need to specify locations to which they are prepared to travel.

The counters table is a housekeeping table to enable a sequence of ascending user numbers.

## Deployment
In the directory in which the app is deployed three subdirectories are needed:
- css/ holds all the css files.
- logs/ where the logfile, speakers.log is saved.
- templates/ where the template files for rendering pages is kept.

Copy the css files and the template files to their relevant directories. Copy the app to the main directory, you will need to change its permissions to execute by executing chmod +x speakers, then to run it enter the command ./speakers

The database MongoDB, will also need to be installed. the MongoDB.com site has very good instructions for installing it on your system.
