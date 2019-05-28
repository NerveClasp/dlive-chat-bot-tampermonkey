const fetch = require('node-fetch');
const fs = require('fs');
var express = require('express');
var app = express();

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
    referrer: 'https://dlive.tv/NerveClasp',
    referrerPolicy: 'no-referrer-when-downgrade',
    body:
      '{"operationName":"LivestreamChatroomInfo","variables":{"displayname":"NerveClasp","isLoggedIn":false,"limit":20},"extensions":{"persistedQuery":{"version":1,"sha256Hash":"c38d67b66455636fee3e0c3f96e5aa53cf344cc99386def56de6b51a27ae36a7"}}}',
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
            newMessages.push(`${sender.displayname}: ${content.trim()}`);
            break;
          case 'Gift':
            break;
          case 'Follow':
            break;
          default:
            break;
        }
      });
      fs.writeFile('chat.txt', messages.join('\n'), err => {
        if (err) console.log(err);
      });
      messages = newMessages;
      // console.log('json', json);
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
