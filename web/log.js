(() => {
	'use strict';
	if (document.location.hash <= 1)
		return;

	var el = document.querySelector(document.location.hash);
	if (el) {
		el.classList.add('highlight');
		setTimeout(() => {
			scrollTo(0, el.offsetTop - window.innerHeight / 2);
		}, 0);
	}

	var hashChanged = () => {
		var el = document.querySelector('div.highlight');
		el.classList.remove('highlight');
		el = document.querySelector(document.location.hash);
		if (el)
			el.classList.add('highlight');
	};
	addEventListener('hashchange', hashChanged);

	document.querySelector('#content').addEventListener('click', (ev) => {
		var target = ev.target;
		if (target.tagName != 'A')
			return;
		ev.preventDefault();
		history.pushState(null, '', target.href);
		hashChanged();
	});
})();
