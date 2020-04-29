from random import randint


def name_generator():
    """
    creates random name using prefix and podstfix list

    Args:
       None

    Returns:
       Random created name

    Raises:
       None

    """
    prefix_list = [
        "admiring", "adoring", "affectionate", "agitated", "amazing", "angry",
        "awesome", "beautiful", "blissful", "bold", "boring", "brave", "busy",
        "charming", "clever", "cool", "compassionate", "competent",
        "confident", "crazy", "dazzling", "determined", "distracted",
        "ecstatic", "elegant", "eloquent", "epic", "exciting", "festive",
        "flamboyant", "focused", "friendly", "frosty", "funny", "gifted",
        "goofy", "gracious", "great", "stoic", "strange", "suspicious",
        "tender", "thirsty", "trusting", "unruffled", "upbeat",	"vibrant",
        "vigilant", "vigorous", "wizardly", "wonderful", "youthful", "zealous",
        "dreamy", "eager", "sweet", "zen"
      ]

    postfix_list = [
         "deepak", "vrushal", "abhishek", "kausthubh", "sandeep", "praveen",
         "shailesh", "sachine", "chetan", "smita", "digvijay", "kedar",
         "prayas", "mandar", "ashish", "arvind"
         ]

    name = prefix_list[randint(0, len(prefix_list)-1)]+"-"+postfix_list[randint(0, len(postfix_list)-1)]+"-"+str(randint(100, 999))
    return name
