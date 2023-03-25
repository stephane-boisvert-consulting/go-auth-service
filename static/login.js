$(document).ready(function() {
  $("#login-form").submit(function(event) {
    // Empêcher la soumission du formulaire
    event.preventDefault();

    // Récupérer les champs de formulaire
    var username = $("#username").val();
    var password = $("#password").val();
    var token = $("#token").val();

    // Envoyer une requête POST au serveur d'authentification
    $.ajax({
      type: "POST",
      url: "/auth",
      data: {
        username: username,
        password: password,
        token: token
      },
      success: function(data) {
        // Rediriger l'utilisateur vers une page protégée
        window.location.href = "/protected";
      },
      error: function(xhr, status, error) {
        // Afficher un message d'erreur
        $("#error-message").text("Erreur de connexion : " + xhr.responseText);
      }
    });
  });
});