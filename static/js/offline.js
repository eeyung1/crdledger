// Wires the offline page's retry button. A separate file (not an inline
// <script>) because the CSP here forbids inline scripts and event
// handlers, on the static offline fallback page same as everywhere else.
document.addEventListener('DOMContentLoaded', function () {
	var btn = document.getElementById('offline-retry');
	if (btn) {
		btn.addEventListener('click', function () {
			location.reload();
		});
	}
});
