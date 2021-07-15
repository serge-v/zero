# Zero

Build and deploy golang application in a second.

## Usage

Add to your application:

	var deploy = flag.Bool("deploy", false, "deploy to zero environment")

	func main() {
		flag.Parse()

		if *deploy {
			if err := zero.Deploy(); err != nil {
				log.Fatal(err)
			}
		}

		println("hello world")	
	}

Run:

	go build && ./myall -deploy

