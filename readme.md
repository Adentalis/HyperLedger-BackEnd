# Hyperledger BackEnd

This project is about testing Hyperledger Fabric from IBM in the SAP Hyperledger Service

The programm is a test car fleet. Its possible to create/read/update/delete cars and users. A user can borrow a car if the car isnt borrowed by someone else yet. After the car is borrowed it can be returned and a TravelLog with all the metadata is automatically created.
Thats the whole point of this system. Automate everything with the blockchain so the user dont have to write data manually on a sheet of paper with all the information about the borrow. 


## How to use
The whole Chaincode (businesslogic) is written in GO.
The interface to interact with the "blockchain" is a REST API. Both files had to be designed and made from scratch in order to get it done.

Its possible to interact directly with the blockchain in the SAP service itself. But thats just lame, so I build a suited Frontend as well for it wich can be found on my account here on GitHub as well :)

The programm exists out of 3 files
The first one is the chaincode.yaml. In it is just the name and version of the programm.
The second and third file have to be  in a folder called "src" in order to be accepted of the SAP service. These two files are the chaincode itself and the REST API interface made with swagger.

For a more detailled explanation you can read the german documentation i wrote in my job. 
It has exactly like the Bitcoin Whitepaper just 9 pages :) #FunFact



