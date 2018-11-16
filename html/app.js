function getCookie(name) {
    function escape(s) { return s.replace(/([.*+?\^${}()|\[\]\/\\])/g, '\\$1'); };
    var match = document.cookie.match(RegExp('(?:^|;\\s*)' + escape(name) + '=([^;]*)'));
    return match ? match[1] : null;
}

new Vue({
    el: '#app',
    
    data: {
        ws: null,
        newMsg: '',
        chatContent: '',
        username: null,
        joined: false
    },

    created: function() {
        //console.log("this was created");
        var self = this;
        this.ws = new WebSocket('ws://' + window.location.host + '/ws');
        this.ws.addEventListener('message', function(e) {
            var msg = JSON.parse(e.data);
            self.chatContent += '<div class="chip">'
                + '<img src="' + self.gravatarURL(msg.email) + '">'
                + msg.username
                + '</div>'
                + emojione.toImage(msg.message) + '<br/>';

            var element = document.getElementById('chat-messages');
            element.scrollTop = element.scrollHeight;
        });
        this.join();
    },

    methods: {
        send: function () {
            if (this.newMsg != '') {
                this.ws.send(
                    JSON.stringify({
                        username: this.username,
                        message: $('<p>').html(this.newMsg).text()
                    }
                ));
            }
        },

        join: function () {
            console.log("joining");

            this.username = getCookie("user");
            console.log(this.username);
            //this.username = $('<p>').html(this.username).text();
            if (this.username != "") {
                this.joined = true;
            }
        },

        gravatarURL: function(email) {
            return 'http://www.gravatar.com/avatar/' + CryptoJS.MD5(email);
        }
    }
});
