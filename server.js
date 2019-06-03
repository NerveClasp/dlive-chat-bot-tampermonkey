const fetch = require('node-fetch');
const fs = require('fs');
const express = require('express');
const app = express();
const config = require('./config.json');

const showLastN = 6;
let followers = {};
let donations = {};
let firstRun = true;
let messages = [];

setInterval(() => {
  fetch('https://graphigo.prd.dlive.tv/', {
    credentials: 'omit',
    headers: {
      accept: '*/*',
      'content-type': 'application/json',
      fingerprint: '',
      gacid: '1409962316.1558770788',
    },
    referrer: `https://dlive.tv/${config.userName}`,
    referrerPolicy: 'no-referrer-when-downgrade',
    body: `{"operationName":"LivestreamChatroomInfo","variables":{"displayname":"${
      config.userName
    }","isLoggedIn":false,"limit":20},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"c38d67b66455636fee3e0c3f96e5aa53cf344cc99386def56de6b51a27ae36a7"}}}`,
    method: 'POST',
    mode: 'cors',
  })
    .then(resp => resp.json())
    .then(json => {
      const { chats } = json.data.userByDisplayName;
      const newMessages = [];
      chats.forEach(c => {
        const { type, content, sender, gift, amount } = c;
        switch (type) {
          case 'Message':
            newMessages.push({
              sender: sender.displayname,
              avatar: sender.avatar,
              message: content,
            });
            break;
          case 'Gift':
            break;
          case 'Follow':
            break;
          default:
            break;
        }
      });
      // skip saving to text file for now
      // fs.writeFile('chat.txt', messages.join('\n'), err => {
      //   if (err) console.log(err);
      // });
      messages = newMessages;
    })
    .catch(err => console.log('err', err));
}, 3000);

app.use(express.static('chat'));

app.get('/', express.static('chat'));

app.get('/api/messages', (req, res) => {
  const lastNMessages = [...messages]
    .reverse()
    .map((m, i) => (i < showLastN ? m : null))
    .filter(m => m !== null);
  res.send(JSON.stringify(lastNMessages));
});

app.listen(3000, function() {
  console.log('Example app listening on port 3000!');
});
