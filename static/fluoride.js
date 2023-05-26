// @license magnet:?xt=urn:btih:90dc5c0be029de84e523b9b3922520e79e0e6f08&dn=cc0.txt CC0

var reverseActions = {
	"like": "unlike",
	"unlike": "like",
	"retweet": "unretweet",
	"unretweet": "retweet"
};

var csrfToken = "";
var antiDopamineMode = false;

function checkCSRFToken() {
	var tag = document.querySelector("meta[name='csrf_token']");
	if (tag)
		csrfToken = tag.getAttribute("content");
}

function checkAntiDopamineMode() {
	var tag = document.querySelector("meta[name='antidopamine_mode']");
	if (tag)
		antiDopamineMode = tag.getAttribute("content") === "true";
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
	f.querySelector("[type='submit']").value = action;
	f.action = "/" + action + "/" + id;
	f.dataset.action = action;
}

function handleLikeForm(id, f) {
	f.onsubmit = function(event) {
		event.preventDefault();

		var action = f.dataset.action;
		var forms = document.
			querySelectorAll(".status-"+id+" .status-like");
		for (var i = 0; i < forms.length; i++) {
			updateActionForm(id, forms[i], reverseActions[action]);
		}

		var body = "csrf_token=" + encodeURIComponent(csrfToken);
		var contentType = "application/x-www-form-urlencoded";
		http("POST", "/fluoride/" + action + "/" + id, 
			body, contentType, function(res, type) {

			if (antiDopamineMode)
				return;
			var data = JSON.parse(res);
			var count = data.data;
			if (count === 0)
				count = "";
			var counts = document.
				querySelectorAll(".status-"+id+" .status-like-count");
			for (var i = 0; i < counts.length; i++) {
				if (count > 0) {
					counts[i].innerHTML = "(" + count + ")";
				} else {
					counts[i].innerHTML = "";
				}
			}
		}, function(err) {
			for (var i = 0; i < forms.length; i++) {
				updateActionForm(id, forms[i], action);
			}
		});
	}
}

function handleRetweetForm(id, f) {
	f.onsubmit = function(event) {
		event.preventDefault();

		var action = f.dataset.action;
		var forms = document.
			querySelectorAll(".status-"+id+" .status-retweet");
		for (var i = 0; i < forms.length; i++) {
			updateActionForm(id, forms[i], reverseActions[action]);
		}

		var body = "csrf_token=" + encodeURIComponent(csrfToken);
		var contentType = "application/x-www-form-urlencoded";
		http("POST", "/fluoride/" + action + "/" + id, 
			body, contentType, function(res, type) {

			if (antiDopamineMode)
				return;
			var data = JSON.parse(res);
			var count = data.data;
			if (count === 0)
				count = "";
			var counts = document.
				querySelectorAll(".status-"+id+" .status-retweet-count");
			for (var i = 0; i < counts.length; i++) {
				if (count > 0) {
					counts[i].innerHTML = "(" + count + ")";
				} else {
					counts[i].innerHTML = "";
				}
			}
		}, function(err) {
			for (var i = 0; i < forms.length; i++) {
				updateActionForm(id, forms[i], action);
			}
		});
	}
}

function isInView(el) {
	var ract = el.getBoundingClientRect();
	if (ract.top > 0 && ract.bottom < window.innerHeight)
		return true;
	return false;
}

function handleReplyToLink(a) {
	if (!a)
		return;
	var id = a.getAttribute("href");
	if (!id || id[0] != "#")
		return;
	a.onmouseenter = function(event) {
		var id = event.target.getAttribute("href");
		var status = document.querySelector(id);
		if (!status)
			return;
		if (isInView(status)) {
			status.classList.add("highlight");
		} else {
			var copy = status.cloneNode(true);
			copy.id = "reply-to-popup";
			var ract = event.target.getBoundingClientRect();
			copy.style["max-width"] = (window.innerWidth - ract.left - 32) + "px";
			if (ract.top > window.innerHeight / 2) {
				copy.style.bottom = (window.innerHeight - 
					window.scrollY - ract.top) + "px";
			}
			event.target.parentElement.appendChild(copy);
		}
	}
	a.onmouseleave = function(event) {
		var popup = document.getElementById("reply-to-popup");
		if (popup) {
			event.target.parentElement.removeChild(popup);    
		} else {
			var id = event.target.getAttribute("href");
			document.querySelector(id)
				.classList.remove("highlight");
		}
	}
}

function handleReplyLink(a) {
	a.onmouseenter = function(event) {
		var id = event.target.getAttribute("href");
		var status = document.querySelector(id);
		if (!status)
			return;
		if (isInView(status)) {
			status.classList.add("highlight");
		} else {
			var copy = status.cloneNode(true);
			copy.id = "reply-popup";
			var ract = event.target.getBoundingClientRect();
			copy.style["max-width"] = (window.innerWidth - 98) + "px";
			if (ract.left > window.innerWidth / 2) {
				copy.style.right = (window.innerWidth -
					ract.right - 12) + "px";
			}
			event.target.parentElement.appendChild(copy);
		}
	}
	a.onmouseleave = function(event) {
		var popup = document.getElementById("reply-popup");
		if (popup) {
			event.target.parentElement.removeChild(popup);
		} else {
			var id = event.target.getAttribute("href");
			document.querySelector(id).classList.remove("highlight");
		}
	}
}

function handleStatusLink(a) {
	if (a.classList.contains("mention"))
		a.removeAttribute("target");
	else
		a.target = "_blank";
}

function setPos(el, cx, cy, mw, mh) {
	var h = el.clientHeight;
	var w = el.clientWidth;
	var left, top;
	if (cx < mw/2) {
		if (w + cx + 20 < mw) {
			left = cx + 20; 
		} else {
			left = (mw - w);
		}
	} else {
		if (cx - w - 20 > 0) {
			left = cx - w - 20; 
		} else {
			left = 0;
		}
	}
	top = (cy - (h/2));
	if (top < 0) {
		top = 0;
	} else if (top + h > mh) {
		top = (mh - h);
	} 
	el.style.left = left + "px";
	el.style.top = top + "px";
}

var imgPrev = null;
var imgX = 0;
var imgY = 0;
function handleImgPreview(a) {
	a.onmouseenter = function(e) {
		var mw = document.documentElement.clientWidth;
		var mh = document.documentElement.clientHeight - 24;
		imgX = e.clientX;
		imgY = e.clientY;
		var img = document.createElement("img");
		img.id = "img-preview";
		img.src = e.target.getAttribute("href");
		img.style["max-width"] = mw + "px";
		img.style["max-height"] = mh + "px";
		imgPrev = img;
		img.onload = function(e2) {
			setPos(imgPrev, imgX, imgY, mw, mh);
		}
		document.body.appendChild(img);
	}
	a.onmouseleave = function(e) {
		var img = document.getElementById("img-preview");
		if (img)
			document.body.removeChild(img);
		imgPrev = null;
	}
	a.onmousemove = function(e) {
		if (!imgPrev)
			return;
		var mw = document.documentElement.clientWidth;
		var mh = document.documentElement.clientHeight - 24;
		imgX = e.clientX;
		imgY = e.clientY;
		setPos(imgPrev, imgX, imgY, mw, mh);
	}
}

function onPaste(e) {
	if (!e.clipboardData.files)
		return;
	var fp = document.querySelector("#post-file-picker")
	var dt = new DataTransfer();
	for (var i = 0; i < fp.files.length; i++) {
		dt.items.add(fp.files[i]);
	}
	for (var i = 0; i < e.clipboardData.files.length; i++) {
		dt.items.add(e.clipboardData.files[i]);
	}
	fp.files = dt.files;
}

document.addEventListener("DOMContentLoaded", function() { 
	checkCSRFToken();
	checkAntiDopamineMode();

	var statuses = document.querySelectorAll(".status-container");
	for (var i = 0; i < statuses.length; i++) {
		var s = statuses[i];
		var id = s.dataset.id;

		var likeForm = s.querySelector(".status-like");
		handleLikeForm(id, likeForm);

		var retweetForm = s.querySelector(".status-retweet");
		handleRetweetForm(id, retweetForm);

		var replyToLink = s.querySelector(".status-reply-to-link");
		handleReplyToLink(replyToLink);

		var replyLinks = s.querySelectorAll(".status-reply-link");
		for (var j = 0; j < replyLinks.length; j++) {
			handleReplyLink(replyLinks[j]);
		}

		var links = s.querySelectorAll(".status-content a");
		for (var j = 0; j < links.length; j++) {
			handleStatusLink(links[j]);
		}
	}

	var links = document.querySelectorAll(".user-profile-decription a, .user-fields a");
	for (var j = 0; j < links.length; j++) {
		links[j].target = "_blank";
	}

	var links = document.querySelectorAll(".status-media-container .img-link");
	for (var j = 0; j < links.length; j++) {
		handleImgPreview(links[j]);
	}

	var pf = document.querySelector(".post-form")
	if (pf)
		pf.addEventListener("paste", onPaste);
});

// @license-end
