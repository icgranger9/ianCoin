# ianCoin
By Ian Granger

### Description ###
Create a cryptocurrency (ianCoin) which allows for creation of new accounts (wallets), transfer of coins between accounts, and has some form of consensus so that consistency can be achieved. To encourage miner involvement, there shall be a reward paid out to miners who create a canonical block. This reward will consist of a certain amount of new coins (to be determined based on the total amount of coins desired), plus the sum of all transaction fees. Transaction fees will be charged as a simple percentage of all transactions, say 5%. Obviously, this is different from other cryptocurrencies, whose fees scale based on the amount of work being done, however these scaling fees are unnecessary on such a simple platform. As far as the transaction itself goes, it will be considered final once there have been three canonical blocks built on top of it. At that time, the funds will become available.

#### Video link ####

[IanCoin Video Overview](https://youtu.be/jetXXRv-OMI "IanCoin")


#### Goals ####
Completed
* Fix any bugs in my code from p1 – p4
* Establish a base form of consensus (Nakamoto Consensus + Proof of Work)
* Add public and private keys, and a way to transfer them between nodes
* create two MPT’s, and add to block: one to store transactions, and one to store wallet balance
* Add base functionality: wallet creations, and transfer between wallets
* Add reward to creation of new canonical block
* Add sum of fees to miner’s reward

In Progress


To Do
* Update PoW, so difficulty scales with number of transactions
* Update transactions to only become final after 3 blocks



#### Usage ####
Requires as args the port number, and the id. For simplicity, I usually use port:99XX, and id: XX, just to keep them straight.

Starts with node1, with Id 01. All other nodes will connect to this on start to get the blockchain

#### Known Bugs ####
* Doesn't stop working on current block if new canonical block is received
