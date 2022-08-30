// args: {operation} {body}

var https = require('https')
var aws4  = require('aws4')

// to illustrate usage, we'll create a utility function to request and pipe to stdout
function request(opts) { https.request(opts, function(res) { res.pipe(process.stdout) }).end(opts.body || '') }


// aws4 will sign an options object as you'd pass to http.request, with an AWS service and region
var opts = {
    host: process.argv[2],
    body: process.argv[3],
    service: 'lambda',
    region: 'us-west-2',
    headers: {
        'Content-Type': 'application/json'
    }
}

// or it can get credentials from process.env.AWS_ACCESS_KEY_ID, etc
aws4.sign(opts)

// we can now use this to query AWS
request(opts)