const http = require('http')
const fs = require('fs');

const args = process.argv.slice(2)
console.log("Initialize", args )

const content = []

function downloadHashes(height = 0){

	if (height == args[2] ){
		return
	}

	http.get('http://'+args[0]+':'+args[1]+'/block-hash?height='+height, (resp) => {
	  let data = '';

	  // A chunk of data_storage has been received.
	  resp.on('data', (chunk) => {
	    data += chunk;
	  });

	  // The whole response has been received. Print out the result.
	  resp.on('end', () => {
	    try{
	    	    json = JSON.parse(data)
		    console.log(json);     	   	    
		    content.push(json)
		    downloadHashes(height+1)
	    }catch(err){
	    	fs.writeFileSync(args[1], content.join("\n") )
	    }	   
	  });

	}).on("error", (err) => {
	  console.log("Error: " + err.message);
	});
	

}

downloadHashes(0)

