var vm1 = new Vue({
    el: '#app',

    data: {
        ws: null, // Our websocket
        newMsg: '', // Holds new messages to be sent to the server
        chatContent: '', // A running list of chat messages displayed on the screen
        email: null, // Email address used for grabbing an avatar
        username: null, // Our username
        joined: false, // True if email and username have been filled in
        payed: false,
        value: 0,
        askforinvoice: 0,
        showModal:false,
    },

    created: function() {
        var self = this;

        this.ws = new WebSocket('ws://' + window.location.host + '/ws');
        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data);
            console.log(msg)
            if (msg.payed == "True") {
                document.getElementById('payed_amout').style.backgroundColor = "darkgreen";
                console.log(msg.value)
                document.title = document.title + " " + msg.value;
                return
            }
            else if (msg.message != "") {
                self.chatContent += '<div class="chip">'
                        + '<img src="' + self.gravatarURL(msg.email) + '">' // Avatar
                        + msg.username
                    + '</div>'
                    + emojione.toImage(msg.message) + '<br/>'; // Parse emojis

                var element = document.getElementById('chat-messages');
                element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
            }
            else if (msg.askforinvoice != "") {
                console.log(msg.askforinvoice);
                document.getElementById('invoice').innerHTML = msg.askforinvoice;
                document.getElementById('barcode').src = "https://api.qrserver.com/v1/create-qr-code/?data=" + msg.askforinvoice + "&amp;size=256x256" 
            }
        });
    },

    methods: {
        send: function () {
            if (this.newMsg != '') {
                
                this.ws.send(
                    JSON.stringify({
                        email: this.email,
                        username: this.username,
                        message: $('<p>').html(this.newMsg).text() // Strip out html
                    }
                ));
                this.newMsg = ''; // Reset newMsg
            }
        },
        join: function () {
            if (!this.email) {
                Materialize.toast('You must enter an email', 2000);
                return
            }
            if (!this.username) {
                Materialize.toast('You must choose a username', 2000);
                return
            }
            this.email = $('<p>').html(this.email).text();
            this.username = $('<p>').html(this.username).text();
            this.joined = true;
        },
        getInvoice: function() {
            this.ws.send(
                JSON.stringify({
                    email: "ereer",
                    askforinvoice: "5",
                    username: "g",
                }
            ));
        },
        gravatarURL: function(email) {
            return 'http://www.gravatar.com/avatar/' + CryptoJS.MD5(email);
        }
    }
});