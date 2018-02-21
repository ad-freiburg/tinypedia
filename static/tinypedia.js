function getParameterByName(name, url) {
    if (!url) url = window.location.href;
    name = name.replace(/[\[\]]/g, "\\$&");
    let regex = new RegExp("[?&]" + name + "(=([^&#]*)|&|#|$)"),
        results = regex.exec(url);
    if (!results) return null;
    if (!results[2]) return '';
    return decodeURIComponent(results[2].replace(/\+/g, " "));
}

function sentenceToHTML(sentence) {
  let sentHTML = $('<span class="sentence">');
  let textRemaining = sentence.text;
  let links = sentence.links;
  if (links) {
    let linkSplits = [];
    for (let i = 0; i < links.length; i++) {
      let link = links[i];
      let start = textRemaining.indexOf(link.text);
      let end = start + link.text.length;

      sentHTML.append(document.createTextNode(textRemaining.substring(0, start)));
      sentHTML.append($('<a />', {
        href: "/?q="+encodeURIComponent(link.page),
        text: link.text
      }));
      textRemaining = textRemaining.substring(end)+" ";
    }
    sentHTML.append(document.createTextNode(textRemaining))
  } else {
    sentHTML.text(textRemaining);
  }


  return sentHTML;
}

function sentencesToHTML(sentences) {
  let sentsHTML = $('<div class="sentences">');
  for (let i = 0; i < sentences.length; i++) {
    sentence = sentences[i];
    sentsHTML.append(sentenceToHTML(sentence));
  }
  return sentsHTML;
}

function sectionsToHTML(sections) {
  let secsHTML = $('<div id="sections">');
  for (let i = 0; i < sections.length; i++) {
    section = sections[i];
    console.log('section: '+section.title)
    if (section.title != "") {
      secsHTML.append($('<h'+(section.depth+1)+'>').text(section.title));
    }
    secsHTML.append(sentencesToHTML(section.sentences));
  }
  return secsHTML;
}

function astToHTML(title, ast) {
  return $('<div id="top">')
    .append(
      $('<h1>').text(title)
    )
    .append(
      sectionsToHTML(ast.sections)
    );
}

$(document).ready(function(){
  let title = getParameterByName('q')
  $.get('wiki/'+encodeURIComponent(title), function(markup){
    ast = wtf.parse(markup)
    /**
     * Handle page redirect's e.g. Moody's ⇒ Moody's Investors Service
     * **/
    if (ast.type == 'redirect') {
      console.log('redirect');
      window.location.replace('?q='+encodeURIComponent(ast.redirect));
      return;
    }
    $('#content').html(
      astToHTML(title, ast)
    );
    $('#debug').html(
      JSON.stringify(ast, null, 2)
    );
  });
})
