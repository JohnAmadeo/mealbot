Goals
- one-shot 
	- one-shot is not a desirable property! what if someone wants to join the group halfway through? need to modify a lot of pairs!! 
- can handle repeat meals when all options are exhausted
- can handle dangling 3 person meals
- can handle forced same-class meals 
	- e.g say we have {A1, A2} & {B1, B2, B3, B4}
	- clearly there will be 2 B year students who will have to be paired every week
	- we can be certain that same-class meals will have to occur for a particular class if the size of the class > the size of the candidate pool (as is the case for B)
- can handle students who join halfway
- what kind of pairings do we want to avoid?
	- naive answer is "repeat pairings", but what if we are in situation where repeat pairing is unavoidable? 

Map
A1 -> {
	min: 0, // min no. of meals A1 has had among all possible candidates
	numMeals: 1, // total no. of meals according to the pairs map
	pairs: { // candidate -> no. meals
		A2: 1 
	}
}

Pseudocode 1
- Construct the map to keep track of current pairs
- For each student X
	- While X has < targetMeals meals scheduled
		- Randomly select a partner Y among all candidates (i.e exclude same-class options)
		- If X & Y have had Z meals, where Z = minMeals, add the pairing to the map for both X & Y
		- Check if minMeals needs to be updated (probably by keeping track of total no. of meals X has had so far) for both X & Y
- Evaluation:
	- Can shut out student (no meals) if odd no. of students 
	- Algorithm cannot accommodate the addition of a student halfway through the semester
		
Pseudocode 2
- Construct a map of maps (conceptually a 2D table)
{
	student1: {
		partner1: int, // num of meals 
		...
	}
}
- For each round (time-based; runs every week)
	- Create empty set to store pairings for this week
	- Set aside a randomly-selected odd student out if needed
	- For each student X
	 	- If already paired (see pairing set), skip
		- Filter out already paired students from candidate list 
		- Select bottom of half least-paired students from the above list 
		- If the list is empty, create a list from same-class students 
		- Randomly select a partner from the list 
		- Record the pairing in the set + pairing memory
	- Pick a random pair and the odd student out (updating memory as well)
		
Upload CSV -> Create members in table 
Select cross-match criteria chip -> Save in org table 
Change rounds (POST all pending rounds) -> Delete all pending rounds rounds; and re-insert
Add/remove member -> add row to Members table 
Run a round -> Load history, run rounds, and insert new history + new pairs		
