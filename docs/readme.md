## Internal docs authoring notes



When you edit the docs in /docs/readme-sync/ and merge to the master branch, you trigger a Travis build stage (readme-sync) that syncs  the Markdown doc files to FullStack public docs at https://docs.developers.optimizely.com/full-stack/docs/optimizely-agent.

### Directory and filename requirements

See https://github.com/flowcommerce/readme-sync. 



### Authoring requirements & limitations

You can author the docs in Github-flavored Markdown, with the following minor restrictions and caveats:

- **links** - You can use relative links, but you have to leave out the .md extension. Like this: [relative link to a doc](./readme-sync/deploy-as-a-microservice). (Future improvement: should be easy to modify readme-sync code to strip out .md extensions if we want working relative links in the markdown stored in github)
- **images** - You can't use relative links. Currently, we use hyperlinks to images stored on the master branch [like this](). 
- **manual edits to updatedAt:** the frontmatter in each markdown page includes an updatedAt field, which you must manually edit when you commit a page, so that the public docs display the correct info at the bottom of the page ("Updated x days ago").
- **no semantic code snippets / language highlighting** ReadMe gets confused if you use a code block snippet that indicates the language. It erratically interprets #code comments as heading markdown syntax. So avoid:

```python
# some python code here
```

and only use:

```
# some code in some unspecified language here  
```

- **no authoring in dash.readme** - If someone doesn't know better, they could edit the Agent docs in dash.readme... but those edits will be overwritten the next time triggers the readme-sync Travis stage. There's no mechanism in dash.readme to warn them not to edit.  Likewise, any suggested edits in ReadMe need to be manually merged to the Github docs rather than merged using ReadMe's mechanism.



### Future improvements

- **preview pages** - we don't have a mechanism to preview the markdown on ReadMe.io before merging to master, so you may want to 'raw copy' your markdown into a hidden ReadMe page to preview your content. (an easy way to achieve previews would be to write a Travis stage that syncs to a private ReadMe sandbox on commits to pull requests.)
-  **build conflict resolution** -  If 2 Travis builds tried to sync to readme at the same time, one could overwrite the other.  Build duration is ~5 minutes. (One way to get to eventual consistency might be to run a nightly Travis build that syncs to Readme.)
- **no automatic updates to readme-sync** - The engine for this sync is an externally developed tool that uses the ReadMe API called readme-sync. We clone readme-sync not from the public repo, but from a mirrored repo (https://travis-ci.com/github/optimizely/readme-sync2) that was set up in early 2020. So over time we'll miss out on updates to readme-sync https://github.com/flowcommerce/readme-sync. 









