## Sample Images

Today I found out that GitHub's raw file viewer on the web does not work on images `16bitsRGB555.bmp` and `16bitsRGB565.bmp`. I have to clarify that both BMP files are formatted just fine, which could be attested by executing `file` command against them on any Linux based machine as follows:
```shell
file 16bitsRGB565.bmp
``` 

The output will be in line of:
```shell
16bitsRGB565.bmp: PC bitmap, Adobe Photoshop with alpha channel mask, 512 x 512 x 16, cbSize 524358, bits offset 70
```

About the repo tag. The tag was increased to v1.0.1 just because of the addition of this README.md file, which may not be viewed as justifiable. But then, users' Go module cache need to catchup and sync with the latest _release_ as it is, especially for anyone who is already using this repo, namely me. :)
