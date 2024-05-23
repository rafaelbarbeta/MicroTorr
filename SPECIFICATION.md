## MicroTorr Project Specification

This markdown document will specify the requirements for this project in depth.
MicroTorr is based in BitTorrent v1.0 specification, available on:
https://wiki.theory.org/BitTorrentSpecification#Message_flow


### .mtorrent file structure

The .mtorrent file structure only has the basic features of a common .torrent file, excluding the optional ones and simplifies the overall structure, minding this simulation. The fields are as follow:

* announce: Has the url of the tracker. In this case, it will be always the only tracker avaliable at: http://tracker-microtorr:8080/announce

* info, which contains the values: length, name, piece length, pieces, id_hash

piece length will be set as 1M for the test file
pieces is a concatenation of all SHA1 hashes 
id_hash is a simplification, it is instead the SHA1 hash of the whole file, which will be used as an identification for this file in the swarm


### Tracker

The Tracker provides just one simple service with the following parameters:

GET /annouce
* info_hash: 20-byte SHA-1 hash of the info dictionary from the .mtorrent file.
In this case, is id_hash.
* peer_id: 20 byte randomly generated id
* ip: peer ip
* port: The port number the peer is listening on.
* event: The event type "started", "stopped", "completed".

Note that many parameters, such as downloaded bytes and remaining bytes are removed from this simplistic implementation

The response is a json that contains only the peers, their listening ports, and ids, such as:
```json
{
    "peers": [
       { "ip": 192.168.10.1, "port": 6881, "peer_id": "aabbccddeeffgghhiijj"},
       { "ip": 192.168.10.1, "port": 6881, "peer_id": "kkllmmnnooppqqrrsstt"},
    ]
} 
```

Peers in MicroTorr will always connect to all other available peers.
Peers are supossed to continue sending GET requests, with the same parameters and event "ping". Peers that do not send a GET request within a time span of 5 minutes, will be considered dead, and removed from peers list. BitTorrent usually set this value to about 30 minutes

### Messages

After connecting to a peer, the client will send a handshake, and wait for its partner's handshake. The Handshake message follow the original implementation format:
* pstrlen: string length of pstr, as a single raw byte (integer 11)
* pstr: string identifier of the protocol, this will be "MICROTORRv1"
* reserved: eight (8) reserved bytes. All current implementations use all zeroes.
* info_hash: 20-byte SHA1 hash of the info key in the metainfo file. This is the same info_hash that is transmitted in tracker requests.
* peer_id: 20-byte string used as a unique ID for the client.

The messages also follow the original specification, with prefix_length, message ID, payload. Prefix length has 4 bytes of size and message ID only has 1. Payload size is message dependant. For the sake of simplicity, choke, unchoke, cancel, interested and not interested message types are stripped from this implementation. The specific behaviour will be discussed in the next topic. Also, a new message "reject" is added. The message type are as follow:

* have len=0005 id=1 pieceindex: Advertise to other peers a newer download piece, so they can update which peers have which pieces
* bitfield len=0001+X, id=2, bitfield: indicates all the pieces a peer has or not. Is sent always when a new connection is made, and only once.
* request: len=009, id=3, index, length: Used to request a block. Note that the field begin does not exist in this implementation
* reject: len=004, index: Used to inform the requesting peer that the block cannot yet be sent. This will cause the other peer to choose another piece with a "lower rarity" until find someone that can sent him the piece. This is a simplification of the original protocol
* piece: len=0005+X, id=7, index, block: The actual block piece.

### Torrent Client

When the program is executed in the command line, it will do the following actions:
1. Read and Parse the .mtorrent file
2. Contact the tracker, and request the peers IPs
3. Estabilish a TCP channel to **all** the peers
4. Send a hanshake, and wait for others handshake and bitfield message
5. Initialize internal structures and routines and start torrenting

Each peer will have a maximum of 1 download and floor(log2(number_of_peers)) of simultaneous uploads at any given time. Peers will always upload a piece when requested, as long it does not exceed the upload limit. If it is not the case, a reject message will be sent. The peer will either choose othe peer that has the same requested piece, or will try to get a lesser rare piece, as described bellow. For the sake of simplicity again, a piece rarity is classified by the number of peers that already downloaded it.

Peers will try to download the "rarest" pieces first, following the BitTorrent protocol. The only exception is when a peer receives a reject message. If the are no other peer with the same piece, or they have also rejected the upload, he will try to get another random piece with "lower rarity", which is a piece with the number of +1 peers who have it, in comparison with the previous requested (and rejected) piece. If the piece is more widespread, the are probably more peers that have "free upload slots" available. He will try this again and again until finding someone, or in the worst scenario, whe will reset the algorithm. This may be very inefficient if the selected peers are always "busy", but the number of slots peer upload is enough for this "micro" implementation.

Peers will also save download delays from other peers in a "delay" structure. In the beginning, the "delay" for each peer is setted to 0, as a way to encourage trying other connections to get a piece, if possible.
In the case more than one peer have a selected piece, the tie is resolved by selected the peer with less known delay from the previous piece downloaded. If the delays are the same, the choice is random. Also, when a peer is rejected, it will save MaxInt as the peer delay value.

This may deviates from the specification of BitTorrent, but again, for the sake of simplicity, this is the way the application will behave

It is expected that the peers will at first download everthing from the seeders, and then will cooperate if they find that downloading from other peers is faster than dowloading from the seeder. 

All received pieces are sorted and checked against its SHA1 checksum to ensure data integrity. All values are stored in RAM, and when all pieces arrive, it is dummped to disc. The client will enter seeder mode immediately, and will only upload from now on.

A client can also be started in seeder mode by specifying the complete file in command line.

