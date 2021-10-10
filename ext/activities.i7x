Printing the name of something (documented at act_pn) is an activity. [0]

Before printing the name of a thing (called the item being printed)
	(this is the make named things mentioned rule):
	if expanding text for comparison purposes, continue the activity;
	now the item being printed is mentioned.

The standard name printing rule is listed last in the for printing the name rulebook.

Printing the plural name of something (documented at act_ppn) is an activity. [1]

Rule for printing the plural name of something (called the item) (this is the standard
	printing the plural name rule):
	say the printed plural name of the item.
The standard printing the plural name rule is listed last in the for printing
the plural name rulebook.

Printing a number of something (documented at act_pan) is an activity. [2]

Rule for printing a number of something (called the item) (this is the standard
	printing a number of something rule):
	say "[listing group size in words] ";
	carry out the printing the plural name activity with the item.
The standard printing a number of something rule is listed last in the for printing
a number rulebook.

Printing room description details of something (documented at act_details) is an activity. [3]
Printing inventory details of something (documented at act_idetails) is an activity. [4]

Listing contents of something (documented at act_lc) is an activity. [5]
The standard contents listing rule is listed last in the for listing contents rulebook.
Grouping together something (documented at act_gt) is an activity. [6]

Writing a paragraph about something (documented at act_wpa) is an activity. [7]

Listing nondescript items of something (documented at act_lni) is an activity. [8]

Printing the name of a dark room (documented at act_darkname) is an activity. [9]
Printing the description of a dark room (documented at act_darkdesc) is an activity. [10]
Printing the announcement of darkness (documented at act_nowdark) is an activity. [11]
Printing the announcement of light (documented at act_nowlight) is an activity. [12]
Printing a refusal to act in the dark (documented at act_toodark) is an activity. [13]

The look around once light available rule is listed last in for printing the
announcement of light.

This is the look around once light available rule:
	try looking.

Constructing the status line (documented at act_csl) is an activity. [14]
Printing the banner text (documented at act_banner) is an activity. [15]

Reading a command (documented at act_reading) is an activity. [16]
Deciding the scope of something (future action) (documented at act_ds) is an activity. [17]
Deciding the concealed possessions of something (documented at act_con) is an activity. [18]
Deciding whether all includes something (future action) (documented at act_all)
	is an activity. [19]
The for deciding whether all includes rules have outcomes it does not (failure) and
	it does (success).
Clarifying the parser's choice of something (future action) (documented at act_clarify)
	is an activity. [20]
Asking which do you mean (future action) (documented at act_which) is an activity. [21]
Printing a parser error (documented at act_parsererror) is an activity. [22]
Supplying a missing noun (documented at act_smn) is an activity. [23]
Supplying a missing second noun (documented at act_smn) is an activity. [24]
Implicitly taking something (documented at act_implicitly) is an activity. [25]

Rule for deciding whether all includes scenery while taking or taking off or
	removing (this is the exclude scenery from take all rule): it does not.
Rule for deciding whether all includes people while taking or taking off or
	removing (this is the exclude people from take all rule): it does not.
Rule for deciding whether all includes fixed in place things while taking or
	taking off or removing (this is the exclude fixed in place things from
	take all rule): it does not.
Rule for deciding whether all includes things enclosed by the person reaching
	while taking or taking off or removing (this is the exclude indirect
	possessions from take all rule): it does not.
Rule for deciding whether all includes a person while dropping or throwing
	or inserting or putting (this is the exclude people from drop all rule):
	it does not.

Rule for supplying a missing noun while an actor smelling (this is the ambient odour rule):
	now the noun is the touchability ceiling of the player.

Rule for supplying a missing noun while an actor listening (this is the ambient sound rule):
	now the noun is the touchability ceiling of the player.

Rule for supplying a missing noun while an actor going (this is the block vaguely going rule):
	say "You'll have to say which compass direction to go in." (A).

The standard implicit taking rule is listed last in for implicitly taking.

Starting the virtual machine (documented at act_startvm) is an activity. [26]

The enable Glulx acceleration rule is listed first in for starting the virtual machine.


Amusing a victorious player (documented at act_amuse) is an activity. [27]

Printing the player's obituary (documented at act_obit) is an activity. [28]
The print obituary headline rule is listed last in for printing the player's obituary.
The print final score rule is listed last in for printing the player's obituary.
The display final status line rule is listed last in for printing the player's obituary.

Handling the final question is an activity. [29]

The print the final question rule is listed in before handling the final question.
The print the final prompt rule is listed in before handling the final question.
The read the final answer rule is listed last in before handling the final question.
The standard respond to final question rule is listed last in for handling the final question.

This is the print the final prompt rule: say "> [run paragraph on]" (A).


Table of Locale Priorities
notable-object (an object)	locale description priority (a number)
--							--
with blank rows for each thing.

To describe locale for (O - object):
	carry out the printing the locale description activity with O.

To set the/-- locale priority of (O - an object) to (N - a number):
	if O is a thing:
		if N <= 0, now O is mentioned;
		if there is a notable-object of O in the Table of Locale Priorities:
			choose row with a notable-object of O in the Table of Locale Priorities;
			if N <= 0, blank out the whole row;
			otherwise now the locale description priority entry is N;
		otherwise:
			if N is greater than 0:
				choose a blank row in the Table of Locale Priorities;
				now the notable-object entry is O;
				now the locale description priority entry is N;

Printing the locale description of something (documented at act_pld) is an activity. [30]

The locale paragraph count is a number that varies.

Before printing the locale description (this is the initialise locale description rule):
	now the locale paragraph count is 0;
	repeat through the Table of Locale Priorities:
		blank out the whole row.

Before printing the locale description (this is the find notable locale objects rule):
	let the domain be the parameter-object;
	carry out the choosing notable locale objects activity with the domain;
	continue the activity.

For printing the locale description (this is the interesting locale paragraphs rule):
	let the domain be the parameter-object;
	sort the Table of Locale Priorities in locale description priority order;
	repeat through the Table of Locale Priorities:
		carry out the printing a locale paragraph about activity with the notable-object entry;
	continue the activity.

For printing the locale description (this is the you-can-also-see rule):
	let the domain be the parameter-object;
	let the mentionable count be 0;
	repeat with item running through things:
		now the item is not marked for listing;
	repeat through the Table of Locale Priorities:
		if the locale description priority entry is greater than 0,
			now the notable-object entry is marked for listing;
		increase the mentionable count by 1;
	if the mentionable count is greater than 0:
		repeat with item running through things:
			if the item is mentioned:
				now the item is not marked for listing;
		begin the listing nondescript items activity with the domain;
		if the number of marked for listing things is 0:
			abandon the listing nondescript items activity with the domain;
		otherwise:
			if handling the listing nondescript items activity with the domain:
				if the domain is the location:
					say "[We] " (A);
				otherwise if the domain is a supporter or the domain is an animal:
					say "On [the domain] [we] " (B);
				otherwise:
					say "In [the domain] [we] " (C);
				if the locale paragraph count is greater than 0:
					say "[regarding the player][can] also see " (D);
				otherwise:
					say "[regarding the player][can] see " (E);
				let the common holder be nothing;
				let contents form of list be true;
				repeat with list item running through marked for listing things:
					if the holder of the list item is not the common holder:
						if the common holder is nothing,
							now the common holder is the holder of the list item;
						otherwise now contents form of list is false;
					if the list item is mentioned, now the list item is not marked for listing;
				filter list recursion to unmentioned things;
				if contents form of list is true and the common holder is not nothing,
					list the contents of the common holder, as a sentence, including contents,
						giving brief inventory information, tersely, not listing
						concealed items, listing marked items only;
				otherwise say "[a list of marked for listing things including contents]";
				if the domain is the location, say " here" (F);
				say ".[paragraph break]";
				unfilter list recursion;
			end the listing nondescript items activity with the domain;
	continue the activity.

Choosing notable locale objects of something (documented at act_cnlo) is an activity. [31]

For choosing notable locale objects (this is the standard notable locale objects rule):
	let the domain be the parameter-object;
	let the held item be the first thing held by the domain;
	while the held item is a thing:
		set the locale priority of the held item to 5;
		now the held item is the next thing held after the held item;
	continue the activity.

Printing a locale paragraph about something (documented at act_plp) is an activity. [32]

For printing a locale paragraph about a thing (called the item)
	(this is the don't mention player's supporter in room descriptions rule):
	if the item encloses the player, set the locale priority of the item to 0;
	continue the activity.

For printing a locale paragraph about a thing (called the item)
	(this is the don't mention scenery in room descriptions rule):
	if the item is scenery, set the locale priority of the item to 0;
	continue the activity.

For printing a locale paragraph about a thing (called the item)
	(this is the don't mention undescribed items in room descriptions rule):
	if the item is undescribed:
		set the locale priority of the item to 0;
	continue the activity.

For printing a locale paragraph about a thing (called the item)
	(this is the set pronouns from items in room descriptions rule):
	if the item is not mentioned, set pronouns from the item;
	continue the activity.

For printing a locale paragraph about a thing (called the item)
	(this is the offer items to writing a paragraph about rule):
	if the item is not mentioned:
		if a paragraph break is pending, say "[conditional paragraph break]";
		carry out the writing a paragraph about activity with the item;
		if a paragraph break is pending:
			increase the locale paragraph count by 1;
			now the item is mentioned;
			say "[conditional paragraph break]";
	continue the activity.

For printing a locale paragraph about a thing (called the item)
	(this is the use initial appearance in room descriptions rule):
	if the item is not mentioned:
		if the item provides the property initial appearance and the
			item is not handled and the initial appearance of the item is
			not "":
			increase the locale paragraph count by 1;
			say "[initial appearance of the item]";
			say "[paragraph break]";
			if a locale-supportable thing is on the item:
				repeat with possibility running through things on the item:
					now the possibility is marked for listing;
					if the possibility is mentioned:
						now the possibility is not marked for listing;
				say "On [the item] " (A);
				list the contents of the item, as a sentence, including contents,
					giving brief inventory information, tersely, not listing
					concealed items, prefacing with is/are, listing marked items only;
				say ".[paragraph break]";
			now the item is mentioned;
	continue the activity.

For printing a locale paragraph about a supporter (called the tabletop)
	(this is the initial appearance on supporters rule):
	repeat with item running through not handled things on the tabletop which
		provide the property initial appearance:
		if the item is not a person and the initial appearance of the item is not ""
			and the item is not undescribed:
			now the item is mentioned;
			say initial appearance of the item;
			say paragraph break;
	continue the activity.

Definition: a thing (called the item) is locale-supportable if the item is not
scenery and the item is not mentioned and the item is not undescribed.

For printing a locale paragraph about a thing (called the item)
	(this is the describe what's on scenery supporters in room descriptions rule):
	if the item is scenery and the item does not enclose the player:
		if a locale-supportable thing is on the item:
			set pronouns from the item;
			repeat with possibility running through things on the item:
				now the possibility is marked for listing;
				if the possibility is mentioned:
					now the possibility is not marked for listing;
			increase the locale paragraph count by 1;
			say "On [the item] " (A);
			list the contents of the item, as a sentence, including contents,
				giving brief inventory information, tersely, not listing
				concealed items, prefacing with is/are, listing marked items only;
			say ".[paragraph break]";
	continue the activity.

For printing a locale paragraph about a thing (called the item)
	(this is the describe what's on mentioned supporters in room descriptions rule):
	if the item is mentioned and the item is not undescribed and the item is
		not scenery and the item does not enclose the player:
		if a locale-supportable thing is on the item:
			set pronouns from the item;
			repeat with possibility running through things on the item:
				now the possibility is marked for listing;
				if the possibility is mentioned:
					now the possibility is not marked for listing;
			increase the locale paragraph count by 1;
			say "On [the item] " (A);
			list the contents of the item, as a sentence, including contents,
				giving brief inventory information, tersely, not listing
				concealed items, prefacing with is/are, listing marked items only;
			say ".[paragraph break]";
	continue the activity.

Issuing the response text of something -- documented at act_resp -- is an
activity on responses. [33]

The standard issuing the response text rule is listed last in for issuing the
response text.