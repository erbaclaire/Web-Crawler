## README

#### Read report.pdf for details on this web crawler project.

### How To Run

**NOTE FOR RUNNING LOCALLY: Run the command ulimit -a | grep open before you run my program. If you have < 1100 for this number then there is a chance the program could crash from too many webpages being open at once. This is controlled by your OS. You can update the max amount of files allowed for a session with different commands on different OS. I did not want to automate a command in the program, though, to change your computing environment so you would have to do this yourself before running the program. You can also just run the program on the cluster with no issue.**

### Disclaimers

1. The web scraping aspect of this project is highly dependent on the structure of the html on the website. Therefore, if the website's underlying html changes, the program may no longer be extracting the correct information from the web pages, which could cause the program to crash. I believe the possibility of this is slim because http://books.toscrape.com/index.html is a non-secured website soley used for the purpose of scraping data. I do not believe anyone is monitoring the site or making updates to the site. This project was created based on the August 10, 2020 version of the website and it was working perfectly with that. If you get errors from the html package, this means the website must have changed. If this is the case then please contact me and I can change the code to be updated for the updated website. This would only take an hour or so, I imagine. -- **If the code does not run, please watch the WATCH_THIS_IF_CODE_DOES_NOT_RUN.mov video in the project folder. This shows me running the test cases to prove to you that it was working and that the only reason it wouldn't be working is if the underlying html changed.**

2. Because the project relies on websites, make sure you have a good Internet connection, especially if you are running the speedup graph code locally. If the Internet goes down the program will crash.

#### Test Cases

Navigate to the “./proj3/src/scraper” directory. 

You can test the sequential version by running the command “go run scraper.go < ./in-files/[input file].txt”. The input file can be the following: genres1, genres2, or genres3. This will print the output (i.e. the cheapest book in each genre) to the command line. These can then visually be compared to the files in ./out-files/answers” directory. 

You can also run “go run scraper.go <  ./in-files/[input file].txt > ./out-files/[input file]_Out.txt” where the input file is one of the following: genres1, genres2, or genres3. Then, you can run the command “go run check.go ./out-files/[input file]_Out.txt ./out-files/answers/[input file]_Out_Ans.txt”. This program will print “Ok!” if the program output matches the answer key or a “FAIL!” message if the output does not match the answer key. See information above about the check.go file in the scraper package for more details.

You can change the command to be “go run scraper.go -p=[numOfThreads]” to execute the parallel version. The input, output, and checking files will be the same.

I would recommend running the following commands/test cases:

* go run scraper.go < ./in-files/genres1.txt > ./out-files/genres1_Out.txt
* go run scraper.go -p=4 < ./in-files/genres2.txt > ./out-files/genres2_Out.txt
* go run scraper.go -p=6 < ./in-files/genres3.txt > ./out-files/genres3_Out.txt

These will produce the results to the ./out-files files.

You can then compare your results to the answer key using the commands:

* go run check.go ./out-files/genres1_Out.txt ./out-files/answers/genres1_Out_Ans.txt
* go run check.go ./out-files/genres2_Out.txt ./out-files/answers/genres2_Out_Ans.txt
* go run check.go ./out-files/genres3_Out.txt ./out-files/answers/genres3_Out_Ans.txt
