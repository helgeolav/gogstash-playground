# hash file

This filter hashes the content of a file.

```json
{
  "filter": [
    {
      "type": "hashfile",
      "field": "file_name",
      "output": "hash",
      "buf_size": 20000,
      "algos": [
        "crc32",
        "md5",
        "sha256",
        "fnv1a",
        "sha512"
      ]
    }
  ]
}
```

The file to hash is found in the event specified by "field". Alogs are the list of hashes to use. If blank then all supported hashes will be used.
The filter will not load if an unknown filter is specified.

The output hash is in its binary form encoded as base64.

Example output:
```json
{
  "hash": {
    "fnv1a": "OObfHVMzKKqDlPRRHpcI3Q==",
    "crc32": "UNbFkg==",
    "md5": "12UcXohM6INPDmCieDpsaQ==",
    "sha256": "5TU5z8hTPQr2yjPaPbWXcvFF7xsUdCBiv/71SCEbynU=",
    "sha512": "R+W447AH2dO4aADxUG2IdyhE5Vg+u4LDHl7pbimYV62FszO+rqkKND+2jQuvzTT1P9lLoJYJZvwcjnL0m8EMZA=="
  }
}
```