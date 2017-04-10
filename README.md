# Gorel
Simple artifacts manager with clean API, encryption and scalable storage

## Usage

Generate a 32 char key for AES encryption/decryption

```bash
# upload artifact
curl -H "Token: <32 random chars>"  --upload-file ./myapp-1.0.0.tar.gz http://gorel.sk5.io/randomfolder/myapp-1.0.0.tar.gz

# download latest artifact
curl -H "Token: <32 random chars>" http://gorel.sk5.io/randomfolder/latest

# download specific artifact
curl -H "Token: <32 random chars>" http://gorel.sk5.io/randomfolder/myapp-0.9.0.tar.gz
```

## Caution

Work in progress, API may break

## Feature

* Simple upload
* AES encryption
* Local Storage
* Latest download

## To do

* [ ] Clean code
* [ ] S3 and GCS backend
* [ ] Automatic artifacts deletion (keep 7 releases)
* [ ] Webhook notification on actions
