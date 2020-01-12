var actionIcons = {
	"like": "/static/icons/star-o.png",
	"dark-like": "/static/icons/dark-star-o.png",
	"unlike": "/static/icons/liked.png",
	"dark-unlike": "/static/icons/liked.png",
	"retweet": "/static/icons/retweet.png",
	"dark-retweet": "/static/icons/dark-retweet.png",
	"unretweet": "/static/icons/retweeted.png",
	"dark-unretweet": "/static/icons/retweeted.png"
};

var reverseActions = {
	"like": "unlike",
	"unlike": "like",
	"retweet": "unretweet",
	"unretweet": "retweet"
};

function http(method, url, success, error) {
	var req = new XMLHttpRequest();
	req.onload = function() {
		if (this.status === 200 && typeof success === "function") {
			success(this.responseText, this.responseType);
		} else if (typeof error === "function") {
			error(this.responseText);
		}
	};
	req.onerror = function() {
		if (typeof error === "function") {
			error(this.responseText);
		}
	};
	req.open(method, url);
	req.send();
}

function updateActionForm(id, f, action) {
	if (Array.from(document.body.classList).indexOf("dark") > -1) {
		f.children[1].src = actionIcons["dark-" + action];
	} else {
		f.children[1].src = actionIcons[action];
	}
	f.action = "/" + action + "/" + id;
	f.dataset.action = action;
}

function handleLikeForm(id, f) {
	f.onsubmit = function(event) {
		event.preventDefault();

		var action = f.dataset.action;
		var forms = document.querySelectorAll(".status-"+id+" .status-like");
		forms.forEach(function(f) {
			updateActionForm(id, f, reverseActions[action]);
		});

		http("POST", "/fluoride/" + action + "/" + id, function(res, type) {
			var data = JSON.parse(res);
			var count = data.data;
			if (count === 0) {
				count = "";
			}
			var counts = document.querySelectorAll(".status-"+id+" .status-like-count");
			counts.forEach(function(c) {
				c.innerHTML = count;
			});
		}, function(err) {
			forms.forEach(function(f) {
				updateActionForm(id, f, action);
			});
		});
	}
}

function handleRetweetForm(id, f) {
	f.onsubmit = function(event) {
		event.preventDefault();

		var action = f.dataset.action;
		var forms = document.querySelectorAll(".status-"+id+" .status-retweet");
		forms.forEach(function(f) {
			updateActionForm(id, f, reverseActions[action]);
		});

		http("POST", "/fluoride/" + action + "/" + id, function(res, type) {
			var data = JSON.parse(res);
			var count = data.data;
			if (count === 0) {
				count = "";
			}
			var counts = document.querySelectorAll(".status-"+id+" .status-retweet-count");
			counts.forEach(function(c) {
				c.innerHTML = count;
			});
		}, function(err) {
			forms.forEach(function(f) {
				updateActionForm(id, f, action);
			});
		});
	}
}

function handleReplyToLink(link) {
	if (!link) {
		return;
	}
	var id = link.firstElementChild.getAttribute('href');
	if (!id || id[0] != '#') {
		return;
	}
	link.onmouseenter = function(event) {
		var id = event.target.firstElementChild.getAttribute('href');
		var status = document.querySelector(id);
		if (!status) {
			return;
		}
		var copy = status.cloneNode(true);
		copy.id = "reply-to-popup";
		link.appendChild(copy);
	}
	link.onmouseleave = function(event) {
		var popup = document.getElementById("reply-to-popup");
		if (popup) {
			event.target.removeChild(popup);    
		}
	}
}

function handleReplyLink(link) {
	link.onmouseenter = function(event) {
		var id = event.target.firstElementChild.getAttribute('href');
		var status = document.querySelector(id);
		if (!status) {
			return;
		}
		var copy = status.cloneNode(true);
		copy.id = "reply-popup";
		link.appendChild(copy);
	}
	link.onmouseleave = function(event) {
		var popup = document.getElementById("reply-popup");
		if (popup) {
			event.target.removeChild(popup);    
		}
	}
}

document.addEventListener("DOMContentLoaded", function() { 
	var statuses = document.querySelectorAll(".status-container");
	statuses.forEach(function(s) {
		var id = s.dataset.id;

		var likeForm = s.querySelector(".status-like");
		handleLikeForm(id, likeForm);

		var retweetForm = s.querySelector(".status-retweet");
		handleRetweetForm(id, retweetForm);

		var replyToLink = s.querySelector(".status-reply-to");
		handleReplyToLink(replyToLink);

		var replyLinks = s.querySelectorAll(".status-reply");
		replyLinks.forEach(handleReplyLink);
	});
});
