# MicroTorr: Peer to Peer file sharing inspired by BitTorrentV1

## About this project
Simplified implementation of a peer to peer (P2P) file sharing network, inspired by BitTorrent V1, written in Go language. Ships with a CLI interface made with Cobra CLI.

This project is capable of running a full simulation of a minimum P2P swarm, including the Tracker server, torrent client, and bencoding a file for sharing with this software. It does implements most of the basic concepts of the BitTorrent protocol, such as:

* Generation of ".torrent" metadata files, with the bencode encoding algorithm
* Peer discovery with Tracker links and timeouts
* Multiple torrent swarms on the same tracker
* "Rarest piece first" download strategy
* Automatic change to seeding mode once download is completed
* Peer selection based on estimated bandwidth and connection speed
* Integrity checking with SHA1 hashing algorithm

On top of that, MicroTorr includes its own features for easy of execution:
* Maximum uploading and downloading speed setting
* Wait for X amount of seeders and Y of leechers before downloading
* Progress Bar for download progress
* Different Levels (and colors) of program verbosity
* Command completion, provided by Cobra CLI
  
## Dependencies and Auxiliary resources
This project is made entirely in Go, and implements most BitTorrent features "from the ground up". Here is a list of third party modules used for building the application:

1. [Cobra Cli](https://github.com/spf13/cobra)
2. [Bencode Encoding](https://github.com/jackpal/bencode-go)
3. [Bandwidth Limiting](https://github.com/conduitio/bwlimit)
4. [Progress Bar](https://github.com/schollz/progressbar)

For reference, this is the specification of the [BitTorrent protocol v1](https://wiki.theory.org/BitTorrentSpecification) this project was inspired from.
  
## Passo a passo
Passos que o grupo realizou para criar, implementar ou projetar o projeto. É importante descrever pelo menos o mais importante para que outras pessoas compreendam como o grupo conseguiu realizar o projeto, quais as atividades feitas, etc, e possam ter meios compreender como reproduzir o projeto, se assim fosse necessário.

Se possível, é legal citar o nome dos arquivos implementados, se forem poucos. Por exemplo, se o seu projeto tiver 4 arquivos, cada um com uma função, citar o nome deles na parte do passo a passo correspondente. Se forem muitos arquivos para uma mesma coisa, não tem problema, podem deixar sem ou deixar apenas o nome da pasta.


## Install

### Install from source

In order to install from source, you need the go compiler. You can follow the steps from the official documentation: [Go Download and install](https://go.dev/doc/install)

Next, clone the repository, install the binary and add the autocomplete script to your shell of choice. Assuming you use the bash shell, these are the commands needed to setup MicroTorr:

```bash
git clone https://github.com/rafaelbarbeta/MicroTorr
cd MicroTorr
go install 
MicroTorr completion bash > microtorr # Check your PATH variable if this fails!
sudo cp microtorr /etc/bash_completion.d/microtorr
```

Restart your shell. Now you can run MicroTorr with autocompletion!

### Install with debian package (Recommended)

Simply download the .deb package in "Releases" and run:
```bash
sudo dpkg -i microtorr.deb
```
This will drop the binary at /usr/local/bin while also configuring autocompletion on bash shell.

No need to install go or anything

## Running MicroTorr

Let's build a simple torrent scenario to explore how MicroTorr locally.
First, let's create a 100M random file with dd command called 'freeware':
```bash
dd if=/dev/urandom of=freeware bs=1M count=100
```
Next, we need to generate a metadata file for freeware
```bash
MicroTorr createMtorr freeware
MicroTorr loadMtorr freware.mtorrent # Checking if .mtorrent was created successfully
```
Once the metadata file is created, we proceed to build a swarm, containing three peers: one is a "seeder" (has the whole file) and two "leechers" (does not have all files). They will discover themselves with help of a Tracker:
```bash
MicroTorr tracker -v 2
```
The terminal will hung. Let's start three new terminals, one for the seeder, and two for the leechers.

Create peer1 and peer2 "home directory":
```bash
mkdir -p peer/peer1
mkdir -p peer/peer2
```

cd to each of them in the different terminals.
Start them by specifying the .mtorrent file alongside with verbosity mode, loopback interfaces and listening ports

Leecher 1
```bash
MicroTorr download ../../freeware.mtorrent -i lo -p 1111 -v 2
```

Leecher 2
```bash
MicroTorr download ../../freeware.mtorrent -i lo -p 2222 -v 2
```

Now you probably see the tracker is receiving "alive" request from both peers. You might notice nothing happened except the "New Connection" messages, because they are waiting for at least one seeder to join the swarm. In this demonstration, we want to limit the seeder bandwidth to 3 MB/s max, so soon or later the leechers might cooperate to achieve a better download speed.

In the seeder terminal, run:
```bash
MicroTorr download freeware.mtorrent -i lo -p 3333 -v 2 -s freeware -u 3000
```

Now watch the progress bar filling. As you seem, the progress bar starts filling up slow and with about 50% of download , it speeds up really quick. 

In the printed statistics, you can see that about half of the file was download from the seeder, and the other half was download from the peer. This is because, following the "rarest piece first" piece download strategy, first each leecher downloads a piece that the minimum of peers currently hold in the swarm. This leads to a interesting behavior, as each leecher will download complimentary pieces first. After that, there will be a time where each piece is owned by the seeder and one of the other peers. Because the other leecher now also holds the remaining pieces of the file, leecher 1, for example, might as well try downloading a piece from the leecher 2. As he can notice, leecher 2 has a much faster upload speed then the seeder, so he will download the remaining pieces from him. The reverse is also true

If you wish to make sure the arrived file is actually the one in the root directory, simply run this command for "freeware" for each "peer directory":

```bash
sha1sum < freeware
```

Hopefully, the output will be the same for all of them. And that's it!

Try with other setups as well. Note that this can also run between different computers on the same LAN, or if you has access to different public IPv4, over the internet!


## Limitations

Please note that this code *does not work as BitTorrent client*, and therefore cannot download torrents from the internet. It is instead a didatic simulation of how the protocol works under the hood. I made it as a way to challenge myself with concurrent programming, networking and of course, the Go language and its ecosystem. Also, I wanted to learn how this protocol actually worked, as I was completely unaware of its inner working until building this code. Bear in mind that some features of the BitTorrent protocol are stripped from this implementation, such as 

* Chocking, and Opportunistic unchocking
* Tit-for-Tat behavior (it does kind of behave that way nevertheless)
* Download and Upload 'slots'
* Retransmission of broken pieces (sha1 detected)
* NAT transversal techniques for peers under NAT
* DHT (Distributed Hash Table)
* Encryption
* Magnet Links

And some others. 
For those reasons, it is not optimized to download really large files over the internet. See the [qbittorrent](https://github.com/qbittorrent/qBittorrent) for this (don't be evil XD)


## Author
* ([Rafael Barbeta](https://github.com/rafaelbarbeta))


## Imagens/screenshots
É necessário colocar pelo menos 3 imagens/screenshots do projeto, porém fiquem a vontade para colocar mais, a medida do que vocês acharem legal para ilustrar o projeto.

Para colocar imagens no Readme do Github, vocês podem usar o seguinte comando (abrir este Readme no modo raw ou como txt):

![Imagem](https://github.com/hackoonspace/Hackoonspace-template/blob/master/exemplo.png)

É preferível que vocês usem imagens hospedadas no próprio GitHub do projeto. É só referenciar o link delas no comando acima.
