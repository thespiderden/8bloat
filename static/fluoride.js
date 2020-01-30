// @license magnet:?xt=urn:btih:90dc5c0be029de84e523b9b3922520e79e0e6f08&dn=cc0.txt CC0

var reverseActions = {
	"like": "unlike",
	"unlike": "like",
	"retweet": "unretweet",
	"unretweet": "retweet"
};

function getCSRFToken() {
	var tag = document.querySelector("meta[name='csrf_token']")
	if (tag)
		return tag.getAttribute("content");
	return "";
}

function http(method, url, body, type, success, error) {
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
	req.setRequestHeader("Content-Type", type);
	req.send(body);
}

function updateActionForm(id, f, action) {
	f.querySelector('[type="submit"]').value = action;
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

		var body = "csrf_token=" + encodeURIComponent(getCSRFToken());
		var contentType = "application/x-www-form-urlencoded";
		http("POST", "/fluoride/" + action + "/" + id, body, contentType, function(res, type) {
			var data = JSON.parse(res);
			var count = data.data;
			if (count === 0) {
				count = "";
			}
			var counts = document.querySelectorAll(".status-"+id+" .status-like-count");
			counts.forEach(function(c) {
				if (count > 0) {
					c.innerHTML = "(" + count + ")";
				} else {
					c.innerHTML = "";
				}
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

		var body = "csrf_token=" + encodeURIComponent(getCSRFToken());
		var contentType = "application/x-www-form-urlencoded";
		http("POST", "/fluoride/" + action + "/" + id, body, contentType, function(res, type) {
			var data = JSON.parse(res);
			var count = data.data;
			if (count === 0) {
				count = "";
			}
			var counts = document.querySelectorAll(".status-"+id+" .status-retweet-count");
			counts.forEach(function(c) {
				if (count > 0) {
					c.innerHTML = "(" + count + ")";
				} else {
					c.innerHTML = "";
				}
			});
		}, function(err) {
			forms.forEach(function(f) {
				updateActionForm(id, f, action);
			});
		});
	}
}

function isInView(el) {
	var ract = el.getBoundingClientRect();
	if (ract.top > 0 && ract.bottom < window.innerHeight) {
		return true;
	}
	return false;
}

function handleReplyToLink(div) {
	if (!div) {
		return;
	}
	var id = div.firstElementChild.getAttribute('href');
	if (!id || id[0] != '#') {
		return;
	}
	div.firstElementChild.onmouseenter = function(event) {
		var id = event.target.getAttribute('href');
		var status = document.querySelector(id);
		if (!status) {
			return;
		}
		if (isInView(status)) {
			status.classList.add("highlight");
		} else {
			var copy = status.cloneNode(true);
			copy.id = "reply-to-popup";
			event.target.parentElement.appendChild(copy);
		}
	}
	div.firstElementChild.onmouseleave = function(event) {
		var popup = document.getElementById("reply-to-popup");
		if (popup) {
			event.target.parentElement.removeChild(popup);    
		} else {
			var id = event.target.getAttribute('href');
			document.querySelector(id)
				.classList.remove("highlight");
		}
	}
}

function handleReplyLink(div) {
	div.firstElementChild.onmouseenter = function(event) {
		var id = event.target.getAttribute('href');
		var status = document.querySelector(id);
		if (!status) {
			return;
		}
		if (isInView(status)) {
			status.classList.add("highlight");
		} else {
			var copy = status.cloneNode(true);
			copy.id = "reply-popup";
			event.target.parentElement.appendChild(copy);
		}
	}
	div.firstElementChild.onmouseleave = function(event) {
		var popup = document.getElementById("reply-popup");
		if (popup) {
			event.target.parentElement.removeChild(popup);    
		} else {
			var id = event.target.getAttribute('href');
			document.querySelector(id)
				.classList.remove("highlight");
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

// @license-end
