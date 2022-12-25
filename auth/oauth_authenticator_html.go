package auth

const LOGGED_IN_PAGE_HTML = `
<!doctype html>
<html lang="en">
  <head>
    <title>UiPath</title>
    <style>
      body {
        margin: 0;
        padding: 0;
      }

      .main {
        background-image: radial-gradient(#ecedee 15%,#f2f2f2 0);
        background-size: 20px 20px;
        height: 100vh;
        width: 100%;
        color: rgba(0,0,0,.87);
        font-size: 14px;
        font-family: Roboto,Noto Sans JP,Noto Sans SC,Helvetica,Arial,sans-serif;
      }

      .content {
        position: fixed;
        top: 40%;
        left: 50%;
        transform: translate(-50%, -50%);
      }

      .icon {
        margin: auto;
        color: #fa4616;
        width: 150px;
      }

      .card {
        margin-top: 25px;
        max-width: 450px;
        background: #fcfefe;
        box-shadow: 0 2px 1px -1px rgb(0 0 0 / 20%), 0 1px 1px 0 rgb(0 0 0 / 14%), 0 1px 3px 0 rgb(0 0 0 / 12%);
        border-radius: 4px;
        padding: 24px;
      }
    </style>
    <script>
      setTimeout(function() {
        window.close()
      }, 5000);      
    </script>
  </head>
  <body>
    <div class="main">
      <div class="content">
        <div class="icon">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 296 110">
            <path d="M0 0h110v110H0V0zm11.4 98.6h87.2V11.4H11.4v87.2zm44-70.2h11.4v33.3c0 15.1-8.5 24.1-23 24.1-14.1 0-22.5-8.9-22.5-24.1V28.4h11.4v33.3c0 8.4 3.5 13.4 11.4 13.4 7.6 0 11.3-4.8 11.3-13.4V28.4zm33 0c0 4-2.9 6.8-7 6.8-4 0-7-2.8-7-6.8 0-4.1 2.9-7 7-7s7 2.8 7 7zM75.8 40.6h11.4v44.7H75.8V40.6zm87.7 6.9c0 12.2-8.1 19.5-20.5 19.5h-10.3v18.4h-11.4v-57H143c12.6 0 20.5 7.3 20.5 19.1zm-11.5 0c0-6.2-3.6-9.9-10.1-9.9h-9.1v20.3h9.1c6.5-.1 10.1-3.7 10.1-10.4zm49.8-6.9h11.4v44.7h-11.4v-5c-3 3.6-7.8 5.6-14.5 5.6-12.1 0-20.7-9.5-20.7-22.9 0-13.2 8.4-23 20.7-23 6.5 0 11.5 2.3 14.5 6.2v-5.6zm0 22.4c0-7.7-4.6-13-11.8-13-7.3 0-11.8 5.1-11.8 13 0 7.4 4 12.9 11.8 12.9 7-.1 11.8-5.1 11.8-12.9zM241 75.6h4.5v9.7h-6c-10.8 0-15.5-5.1-15.5-15.7V50.2h-5.3v-9.6h5.3V28.4h11.4v12.2h10v9.6h-10v19.5c0 3.9 1.2 5.9 5.6 5.9zm53.6-16.3v26h-11.4V60.6c0-6.8-3.5-11.2-10.2-11.2-6.7 0-10.7 4.6-10.7 12.2v23.7h-11.4V25.6h11.4v19.6c2.8-3.4 7.3-5.2 13.5-5.2 10.6-.1 18.8 7.6 18.8 19.3z" fill="currentColor">
            </path>
          </svg>
        </div>
        <div class="card">Successfully logged in, you can use the <code><b>uipathcli</b></code> now</div>
      </div>
    </div>
  </body>
</html>
`
