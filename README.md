# feeds-merge


This actions only support merge multi-feeds into one feed

We can also upload feeds.xml to s3 or something like that.


```yaml
jobs:
  upload:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: 91go/feeds-merge@v1
        with:
          FEEDS_PATH: .github/feeds.yml
          CLIENT_TIMEOUT: 50
          AUTHOR_NAME: xxx
          FEED_LIMIT: 300
      - uses: shallwefootball/s3-upload-action@master
        with:
          aws_key_id: ${{ secrets.AWS_KEY_ID }}
          aws_secret_access_key: ${{ secrets.AWS_SECRET_ACCESS_KEY}}
          aws_bucket: ${{ secrets.AWS_BUCKET }}
          source_dir: 'dirname'
```

Or directly commit changes

```yaml
jobs:
  upload:
    runs-on: ubuntu-latest
    steps:
      - uses: 91go/feeds-merge@v1
      - name: Commit files
        run: |
          git config --local user.email "41898282+github-actions[bot]@users.noreply.github.com"
          git config --local user.name "github-actions[bot]"
          git commit -a -m "Add changes"
      - name: Push changes
        uses: ad-m/github-push-action@master
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          branch: ${{ github.ref }}
```



