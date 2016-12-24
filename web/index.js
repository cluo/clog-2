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
	document.querySelector('select[name=channel]').value = query['channel'];
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

	var renderResults = (hits) => {
		var el = document.querySelector('#hits');
		for (var i = 0; i < hits.length; i++) {
			var hit = hits[i];

			var link = document.createElement('a');
			var href = '/log/' + query['channel'] + '/' + hit['date']
			if (hit['line'] != -1)
				href += '#l' + hit['line'];
			link.href = href;
			link.innerText = hit['date'] + ':' + hit['line'];


			var div = document.createElement('div');
			div.classList.add('hit');
			div.appendChild(link);
			if (hit['line'] != -1) {
				div.appendChild(document.createElement('br'));
				div.appendChild(document.createTextNode(hit['text']));
			}
			el.appendChild(div);
		}
	};
})();
