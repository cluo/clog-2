(() => {
	'use strict';
	if (document.location.hash <= 1)
		return;

	var el = document.querySelector(document.location.hash);
	el.classList.add('highlight');
	setTimeout(() => {
		scrollTo(0, el.offsetTop - window.innerHeight / 2);
	}, 0);
})();
