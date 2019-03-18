window.addEventListener('load', function () {
    document.getElementById('sign-out').onclick = function () {
        firebase.auth().signOut();
    };

    // FirebaseUI config.
    var uiConfig = {
        signInSuccessUrl: '/',
        signInOptions: [
            firebase.auth.GoogleAuthProvider.PROVIDER_ID,
        ],
        tosUrl: '<your-tos-url>'
    };

    firebase.auth().onAuthStateChanged(function (user) {
        if (user) {
            document.getElementById('sign-out').hidden = false;
            document.getElementById('post-form').hidden = false;
            document.getElementById('account-details').textContent =
                'Signed in as ' + user.displayName + ' (' + user.email + ')';
            user.getIdToken().then(function (accessToken) {
                // Add the token to the post form. The user info will be extracted from the token by the server.
                document.getElementById('token').value = accessToken;
            });
        } else {
            var ui = new firebaseui.auth.AuthUI(firebase.auth());
            // Show the Firebase login button.
            ui.start('#firebaseui-auth-container', uiConfig);
            // Update the login state indicators.
            document.getElementById('sign-out').hidden = true;
            document.getElementById('post-form').hidden = true;
            document.getElementById('account-details').textContent = '';
        }
    }, function (error) {
        console.log(error);
        alert('Unable to log in: ' + error)
    });
});