using System.Collections.Generic;
using Pulumi;
using Aws = Pulumi.Aws;

return await Deployment.RunAsync(() => 
{
    var siteBucket = new Aws.S3.Bucket("siteBucket");

    var testFileAsset = new Aws.S3.BucketObject("testFileAsset", new()
    {
        Bucket = siteBucket,
        Source = new FileAsset("file.txt"),
    });

    var testStringAsset = new Aws.S3.BucketObject("testStringAsset", new()
    {
        Bucket = siteBucket,
        Source = new StringAsset("<h1>File contents</h1>"),
    });

    var testRemoteAsset = new Aws.S3.BucketObject("testRemoteAsset", new()
    {
        Bucket = siteBucket,
        Source = new RemoteAsset("https://pulumi.test"),
    });

    var testFileArchive = new Aws.S3.BucketObject("testFileArchive", new()
    {
        Bucket = siteBucket,
        Source = new FileArchive("file.tar.gz"),
    });

    var testRemoteArchive = new Aws.S3.BucketObject("testRemoteArchive", new()
    {
        Bucket = siteBucket,
        Source = new RemoteArchive("https://pulumi.test/foo.tar.gz"),
    });

    var testAssetArchive = new Aws.S3.BucketObject("testAssetArchive", new()
    {
        Bucket = siteBucket,
        Source = new AssetArchive(new Dictionary<string, AssetOrArchive>
        {
            ["file.txt"] = new FileAsset("file.txt"),
            ["string.txt"] = new StringAsset("<h1>File contents</h1>"),
            ["remote.txt"] = new RemoteAsset("https://pulumi.test"),
            ["file.tar"] = new FileArchive("file.tar.gz"),
            ["remote.tar"] = new RemoteArchive("https://pulumi.test/foo.tar.gz"),
            [".nestedDir"] = new AssetArchive(new Dictionary<string, AssetOrArchive>
            {
                ["file.txt"] = new FileAsset("file.txt"),
                ["string.txt"] = new StringAsset("<h1>File contents</h1>"),
                ["remote.txt"] = new RemoteAsset("https://pulumi.test"),
                ["file.tar"] = new FileArchive("file.tar.gz"),
                ["remote.tar"] = new RemoteArchive("https://pulumi.test/foo.tar.gz"),
            }),
        }),
    });

});
