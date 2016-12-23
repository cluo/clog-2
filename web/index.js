(() => {
	'use strict';
	if (window.location.search.length <= 1)
		return;

	var query = {};
	var split = window.location.search.substr(1).split('&');
	for (var i in split) {
		var pair = split[i].split('=');
		query[pair[0]] = decodeURIComponent(pair[1].replace(/\+/g, ' '));
	}
	document.querySelector('input[name=channel]').value = query['channel'];
	document.querySelector('input[name=q]').value = query['q'];

	var xhr = new XMLHttpRequest();
	xhr.onreadystatechange = () => {
		if (xhr.readyState !== XMLHttpRequest.DONE)
			return;
		var data = JSON.parse(xhr.responseText);
		renderResults(data);
	};
	xhr.open('GET', '/search' + window.location.search);
	xhr.send();

	var renderResults = (data) => {
		var hits = data['hits'];
		var el = document.querySelector('#hits');
		for (var i = 0; i < hits.length; i++) {
			var hit = hits[i];
			var link = document.createElement('a');
			var channel = hit['index'].substr(6); // strip "bleve/"
			var split = hit['id'].split(':');
			link.href = '/log/' + channel + '/' + split[0];
			link.innerText = hit['id'];
			var div = document.createElement('div');
			div.appendChild(link)
			el.appendChild(div);
		}
	};
})();
